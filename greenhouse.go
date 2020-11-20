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
	"github.com/stianeikeland/go-rpio"
)

// TODO: change all pins to actual RPIO pins

// ghFile contains the json filename for storing the greenhouses and the components.
const ghFile = "greenhouses.json"

// configFile contains the json filename for storing the configuration of the program itself,
// such as tresholds, mail, etc.
const configFile = "config.json"

// configFolder is the folder where the config files (ghFile and configFile) are stored
const configFolder = "config"

var mu sync.Mutex
var tpl *template.Template
var fm = template.FuncMap{"fdateHM": hourMinute}
var g = *Greenhouse{}

type config = struct {
	RefreshRate time.Duration // Refresh rate for website
}

// Led represents a LED light in the greenhouse.
type Led struct {
	Id     string // e.g. "Main" or "01"
	Active bool
	Pin    rpio.Pin
	Start  time.Time
	End    time.Time
}

// A MoistSensor represents a sensor that measures the soil moisture.
type MoistSensor struct {
	Id      string
	Value   int
	Channel int
}

// A Greenhouse represents a greenhouse with plants, moisture sensors and LED lights.
type Greenhouse struct {
	Id          string
	MoistSs     []*MoistSensor
	Leds        []*Led
	MoistMin    int           // Minimal value for triggering
	MoistValue  int           // Last measured value
	MoistTiming time.Time     // Timing when last measured
	MoistFreq   time.Duration // Frequency for checking moisture
}

func init() {
	//Loading gohtml templates
	tpl = template.Must(template.New("").Funcs(fm).ParseGlob("./templates/*"))

	// Check if config folder exists, else create
	if _, err := os.Stat(configFolder); os.IsNotExist(err) {
		os.Mkdir(configFolder, 4096)
	}
}

func main() {
	log.Println("--------Start of program--------")

	// Load general config, including webserver
	err := loadConfig("./"+configFolder+"/"+configFile, &config)
	checkErr(err)

	// Load greenhouse
	err = loadConfig("./"+configFolder+"/"+ghFile, &gx)

	// Connecting to rpio Pins
	rpio.Open()
	defer rpio.Close()

	// Launch all configured Greenhouses
	log.Printf("There is/are %v greenhouse(s) configured", len(gx))
	for _, g := range gx {
		log.Printf("Greenhouse %s has %v box(es) configured", g.Id, len(g.Boxes))
		// For each box...
		for _, b := range g.Boxes {
			if len(b.MoistSs) != 0 {
				// Monitor moisture
				go b.monitorMoist()
			}
			// Reset start/end time and monitor light for each LED
			for _, l := range b.Leds {
				l.Start = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), l.Start.Hour(), l.Start.Minute(), 0, 0, time.Now().Location())
				l.End = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), l.End.Hour(), l.End.Minute(), 0, 0, time.Now().Location())
				go l.monitorLed()
			}
		}
		// Monitor temperature for all sensors in the Greenhouse
		if len(g.TempSs) != 0 {
			go g.monitorTemp()
		}
	}
	for {
	}
	// log.Println("Launching website...")
	// http.HandleFunc("/", handlerMain)
	// http.Handle("/favicon.ico", http.NotFoundHandler())
	// http.HandleFunc("/stop", handlerStop)
	// log.Fatal(http.ListenAndServe(":8081", nil))
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

func handlerStop(w http.ResponseWriter, req *http.Request) {
	// TODO: rewrite or remove
	log.Println("Executing a hard stop...")
	os.Exit(3)
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
		// TODO: add error handling in case value is zero
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

// loadConfig loads configuration from the given fname (including folder)
// into i interface.
func loadConfig(fname string, i interface{}) error {
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
			// TODO: print it is too low(!)
			mu.Unlock()
			// TODO: start pump for t seconds
			mu.Lock()
			// log.Printf("Pump %s has run for %s in Box %s\n", b.Pump.Id, b.Pump.Dur, b.Id)
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
