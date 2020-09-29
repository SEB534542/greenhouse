package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/SEB534542/seb"
	"github.com/stianeikeland/go-rpio/v4"
)

// TODO: change all pins to actual RPIO pins

// ghFile contains the json filename for storing the greenhouse config.
const ghFile = "greenhouses.json"
const configFile = "config.json"
const configFolder = "config"

var mu sync.Mutex
var tpl *template.Template
var fm = template.FuncMap{"fdateHM": hourMinute}
var gx = []*Greenhouse{}

var config = struct {
	TempCheck   time.Duration // Frequency for checking moisture
	MoistCheck  time.Duration // Frequency for checking temperature
	RefreshRate time.Duration // Refresh rate for website
}{
	time.Second * 10, // Default value
	time.Second * 10, // Default value
	time.Second * 10, // Default value
}

// An Led represents the a LED light in the greenhouse
type Led struct {
	Id     string // e.g. "Main" or "01"
	Active bool
	Pin    rpio.Pin
	Start  time.Time
	End    time.Time
}

// A pump represents the waterpump that can be activated through the Pin to
// add water to the greenhouse.
type Pump struct {
	Id  string `json:"PumpId"`
	Pin rpio.Pin
	Dur time.Duration
}

type MoistSensor struct {
	Id    string
	Value int
	Pin   rpio.Pin
}

// A servo represents a servo motor to open a window for ventilation.
type Servo struct {
	Id   string
	Pin  rpio.Pin
	Open bool
}

// A TempSensor represents a sensor that measures the temperature.
type TempSensor struct {
	Id    string
	Value int
	Pin   rpio.Pin
}

// A Box represents a greenbox with plants in a greenhouse,
// with it's own water pump and moisture sensors.
type Box struct {
	Id         string
	MoistSs    []*MoistSensor
	Leds       []*Led
	MoistMin   int
	MoistValue int
	Pump
}

// A Greenhouse represents a greenhouse consisting of a/multiple box(es)
// with plants, sensors and LED lights.
type Greenhouse struct {
	Id        string
	Servos    []*Servo
	TempSs    []*TempSensor
	Boxes     []*Box
	TempMin   int
	TempMax   int
	TempValue int
}

func init() {
	//Loading gohtml templates
	tpl = template.Must(template.New("").Funcs(fm).ParseGlob("./templates/*"))

	// Check if config folder exists
	if _, err := os.Stat(configFolder); os.IsNotExist(err) {
		os.Mkdir(configFolder, 4096)
	}
}

func main() {
	log.Println("--------Start of program--------")

	// Load config
	loadConfig := func(fname string, i interface{}) error {
		if _, err := os.Stat(fname); os.IsNotExist(err) {
			log.Printf("File '%v' does not exist, creating blank", fname)
			seb.SaveToJSON(i, fname)
		} else {
			data, err := ioutil.ReadFile(fname)
			if err != nil {
				return fmt.Errorf("%s is corrupt. Please delete the file (%v)", fname, err)
			}
			err = json.Unmarshal(data, i)
			if err != nil {
				return fmt.Errorf("%s is corrupt. Please delete the file (%v)", fname, err)
			}
		}
		return nil
	}
	// General config
	err := loadConfig("./"+configFolder+"/"+configFile, &config)
	checkErr(err)
	// Greenhouse config
	err = loadConfig("./"+configFolder+"/"+ghFile, &gx)

	// Connecting to rpio Pins
	rpio.Open()
	defer rpio.Close()
	//	for _, pin := range []rpio.Pin{s1.pinDown, s1.pinUp} {
	//		pin.Output()
	//		pin.High()
	//	}

	// Launch all configured Greenhouses
	log.Printf("There is/are %v greenhouse(s) configured", len(gx))
	for _, g := range gx {
		log.Printf("Greenhouse %s has %v box(es) configured", g.Id, len(g.Boxes))
		// For each box...
		for _, b := range g.Boxes {
			// ... Monitor moisture
			go b.monitorMoist()
			// ... Reset start/end time and monitor light for each LED
			for _, l := range b.Leds {
				l.Start = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), l.Start.Hour(), l.Start.Minute(), 0, 0, time.Now().Location())
				l.End = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), l.End.Hour(), l.End.Minute(), 0, 0, time.Now().Location())
				go l.monitorLed()
			}
		}

		// Monitor temperature for all sensors in the Greenhouse
		go g.monitorTemp()
	}

	log.Println("Launching website...")
	http.HandleFunc("/", handlerMain)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handlerMain(w http.ResponseWriter, req *http.Request) {
	data := struct {
		Time        string
		RefreshRate int
		Gx          []*Greenhouse
	}{
		time.Now().Format("_2 Jan 06 15:04:05"),
		int(config.RefreshRate.Seconds()),
		gx,
	}
	tpl.ExecuteTemplate(w, "index.gohtml", data)
}

// MonitorTemp monitors the temperature in the Greenhouse
// and moves the servo motors accordinly to open or close the window(s).
func (g *Greenhouse) monitorTemp() {
	for {
		values := []int{}
		mu.Lock()
		for _, s := range g.TempSs {
			s.getTemp()
			values = append(values, s.Value)
		}
		g.TempValue = calcAverage(values...)
		log.Printf("Average temperature in %s: %v degrees based on: %v", g.Id, g.TempValue, values)

		// Evaluating Temperature and moving window(s) accordingly
		switch {
		case g.TempValue > g.TempMax:
			log.Printf("Too hot, opening window(s) for greenbox %s...", g.Id)
			for _, s := range g.Servos {
				mu.Unlock()
				s.unshut()
				mu.Lock()
			}
		case g.TempValue < g.TempMin:
			log.Printf("Too cold, closing window(s) for greenbox %s...", g.Id)
			for _, s := range g.Servos {
				mu.Unlock()
				s.shut()
				mu.Lock()
			}
		}
		log.Printf("Snoozing monitorTemp for %v seconds", config.TempCheck.Seconds())
		for i := 0; i < int(config.TempCheck.Seconds()); i++ {
			mu.Unlock()
			time.Sleep(time.Second)
			mu.Lock()
		}
		mu.Unlock()
	}
}

// GetTemp retrieves the temperature from the Temperature Sensor.
func (s *TempSensor) getTemp() {
	// TODO: get Moist value from RPIO
	// Seed the random number generator using the current time (nanoseconds since epoch)
	rand.Seed(time.Now().UnixNano())

	// Much harder to predict...but it is still possible if you know the day, and hour, minute...
	s.Value = rand.Intn(30)
}

// Unshut opens the window through the servo motor.
func (s Servo) unshut() {
	if s.Open == false {
		s.move()
	}
}

// Shut closes the window through the servo motor.
func (s Servo) shut() {
	if s.Open == true {
		s.move()
	}
}

// Move either opens or closes the window,
// depending if the window is Open (true) or not (false).
func (s Servo) move() {
	if s.Open == true {
		log.Println("Opening window...")
	} else {
		log.Println("Closing window...")
	}
}

// MonitorLED checks if LED should be switched on or off.
func (l *Led) monitorLed() {
	for {
		mu.Lock()
		switch {
		case time.Now().After(l.End):
			log.Println("Resetting Start and End to tomorrow for LED", l.Id)
			l.Start = l.Start.AddDate(0, 0, 1)
			l.End = l.End.AddDate(0, 0, 1)
			fallthrough
		case time.Now().Before(l.Start):
			log.Printf("Turning LED %s off for %v sec until %s...", l.Id, int(time.Until(l.Start).Seconds()), l.Start.Format("02-01 15:04"))
			l.switchLedOff()
			for i := 0; i < int(time.Until(l.Start).Seconds()); i++ {
				mu.Unlock()
				time.Sleep(time.Second)
				mu.Lock()
			}
			fallthrough
		case time.Now().After(l.Start) && time.Now().Before(l.End):
			log.Printf("Turning LED %s on for %v sec until %s", l.Id, int(time.Until(l.End).Seconds()), l.End.Format("02-01 15:04"))
			l.switchLedOn()
			for i := 0; i < int(time.Until(l.End).Seconds()); i++ {
				mu.Unlock()
				time.Sleep(time.Second)
				mu.Lock()
			}
		}
		mu.Unlock()
	}
}

// SwitchLedOn switches the LED on.
func (l *Led) switchLedOn() {
	if !l.Active {
		l.switchLed()
	}
	return
}

// SwitchLedOn switches the LED off.
func (l *Led) switchLedOff() {
	if l.Active {
		l.switchLed()
	}
	return
}

// SwitchLed switches the LED on or off,
// depening if the LED is active (true) or not (false).
func (l *Led) switchLed() {
	if l.Active {
		// TODO: turn LED off
		//log.Printf("Turning Led %s off...", l.Id)
		l.Active = false
	} else {
		// TODO: turn LED on
		//log.Printf("Turning Led %s on...", l.Id)
		l.Active = true
	}
}

// MonitorMoist monitors moisture and if insufficent enables the waterpump.
func (b *Box) monitorMoist() {
	for {
		values := []int{}
		mu.Lock()
		for _, s := range b.MoistSs {
			s.getMoist()
			values = append(values, s.Value)
		}
		b.MoistValue = calcAverage(values...)
		log.Printf("Average moisture in box %v: %v based on: %v", b.Id, b.MoistValue, values)
		if b.MoistValue <= b.MoistMin {
			mu.Unlock()
			// TODO: start pump for t seconds
			mu.Lock()
			log.Printf("Pump %s has run for %s in Box %s\n", b.Pump.Id, b.Pump.Dur, b.Id)
		}
		log.Printf("Snoozing MonitorMoist for %v seconds", config.MoistCheck.Seconds())
		for i := 0; i < int(config.MoistCheck.Seconds()); i++ {
			mu.Unlock()
			time.Sleep(time.Second)
			mu.Lock()
		}
		mu.Unlock()
	}
}

// GetMoist gets retrieves the current moisture value from the sensor
// and stores it in MoistSensor.Value.
func (s *MoistSensor) getMoist() {
	// TODO: get Moist value from RPIO
	// Seed the random number generator using the current time (nanoseconds since epoch)
	rand.Seed(time.Now().UnixNano())

	// Much harder to predict...but it is still possible if you know the day, and hour, minute...
	s.Value = rand.Intn(1000)
}

// CheckErr evaluates err for errors (not nil)
// and triggers a log.Panic containing the error.
func checkErr(err error) {
	if err != nil {
		log.Panic("Error:", err)
	}
	return
}

// HourMinute returns a variable of time.Time as a string in format "15:04"
// This function is used for displaying time on a gohtml webpage.
func hourMinute(t time.Time) string {
	return t.Format("15:04")
}

// CalcAverage takes a variadic parameter of integers and
// returns the average integer.
func calcAverage(xi ...int) int {
	total := 0
	for _, v := range xi {
		total = total + v
	}
	return total / len(xi)
}
