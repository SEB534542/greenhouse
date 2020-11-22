package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/SEB534542/seb"
	"github.com/stianeikeland/go-rpio"
)

// TODO: change all pins to actual RPIO pins

// ghFile contains the json filename for storing the greenhouses and the components.
const ghFile = "greenhouse.json"

// configFile contains the json filename for storing the configuration of the program itself,
// such as tresholds, mail, etc.
const configFile = "config.json"

// configFolder is the folder where the config files (ghFile and configFile) are stored
const configFolder = "config"

// moistFile is the file where moisture data is stored
const moistFile = "moisture_stats.csv"

var mu sync.Mutex
var tpl *template.Template
var fm = template.FuncMap{"fdateHM": hourMinute, "fsec": seconds}
var g = &Greenhouse{}
var c = Config{}

type Config struct {
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

// A SoilSensor represents a sensor that measures the soil moisture.
type SoilSensor struct {
	Id      string
	Channel int
	Value   int
	Time    time.Time
}

// A Greenhouse represents a greenhouse with plants, moisture sensors and LED lights.
type Greenhouse struct {
	Id          string
	Led         *Led
	SoilSensors []*SoilSensor
	MoistMin    int           // Minimal value for triggering
	MoistValue  int           // Last measured value
	MoistTime   time.Time     // Timing when last measured
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
	err := loadConfig("./"+configFolder+"/"+configFile, &c)
	checkErr(err)

	// Load greenhouse
	err = loadConfig("./"+configFolder+"/"+ghFile, &g)

	// Connecting to rpio Pins
	rpio.Open()
	defer rpio.Close()

	// Launch Greenhouse
	if g.Led != nil {
		// Reset start/end time and monitor light for  LED
		g.Led.Start = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), g.Led.Start.Hour(), g.Led.Start.Minute(), 0, 0, time.Now().Location())
		g.Led.End = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), g.Led.End.Hour(), g.Led.End.Minute(), 0, 0, time.Now().Location())
		g.Led.Pin.Output()
		go func() {
			for {
				g.Led.monitorLed()
			}
		}()
	}

	// Monitor moisture
	if len(g.SoilSensors) != 0 {
		go func() {
			for {
				g.monitorSoil()
				log.Printf("Next soil measurement is in %v at %v", g.SoilFreq, g.SoilTime.Add(g.SoilFreq).Format("15:04"))
				for time.Until(g.SoilTime.Add(g.SoilFreq)) > 0 {
					time.Sleep(time.Second)
				}
			}
		}()
	}

	log.Println("Launching website...")
	http.HandleFunc("/", handlerMain)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.HandleFunc("/toggleled", handlerToggleLed)
	http.HandleFunc("/soilcheck", handlerSoilCheck)
	http.HandleFunc("/stop", handlerStop)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handlerMain(w http.ResponseWriter, req *http.Request) {
	mu.Lock()
	stats := seb.ReverseXss(readCSV(moistFile))
	data := struct {
		Time string
		Config
		*Greenhouse
		Stats [][]string
	}{
		time.Now().Format("_2 Jan 06 15:04:05"),
		c,
		g,
		stats,
	}
	mu.Unlock()
	tpl.ExecuteTemplate(w, "index.gohtml", data)
}

func handlerToggleLed(w http.ResponseWriter, req *http.Request) {
	g.Led.toggleLed()
	http.Redirect(w, req, "/", http.StatusFound)
}

func handlerSoilCheck(w http.ResponseWriter, req *http.Request) {
	g.monitorSoil()
	http.Redirect(w, req, "/", http.StatusFound)
}

func handlerStop(w http.ResponseWriter, req *http.Request) {
	g.Led.switchLedOff()
	rpio.Close()
	log.Println("Shutting down...")
	os.Exit(3)
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

// MonitorLED checks if LED should be switched on or off.
func (l *Led) monitorLed() {
	mu.Lock()
	switch {
	case time.Now().After(l.End):
		log.Println("Resetting Start and End to tomorrow for LED", l.Id)
		l.Start = l.Start.AddDate(0, 0, 1)
		l.End = l.End.AddDate(0, 0, 1)
		fallthrough
	case time.Now().Before(l.Start):
		log.Printf("Turning LED %s off and snoozing for %v until %s...", l.Id, time.Until(l.Start), l.Start.Format("02-01 15:04"))
		l.switchLedOff()
		for time.Until(l.Start) > 0 {
			mu.Unlock()
			time.Sleep(time.Second)
			mu.Lock()
		}
		fallthrough
	case time.Now().After(l.Start) && time.Now().Before(l.End):
		log.Printf("Turning LED %s on and snoozing for %v until %s...", l.Id, time.Until(l.End), l.End.Format("02-01 15:04"))
		l.switchLedOn()
		for time.Until(l.End) > 0 {
			mu.Unlock()
			time.Sleep(time.Second)
			mu.Lock()
		}
	}
	mu.Unlock()
}

// SwitchLedOn switches LED on.
func (l *Led) switchLedOn() {
	l.Pin.Write(rpio.Low)
	l.Active = true
	return
}

// SwitchLedOn switches LED off.
func (l *Led) switchLedOff() {
	l.Pin.Write(rpio.High)
	l.Active = false
	return
}

func (l *Led) toggleLed() {
	if l.Active == true {
		l.switchLedOff()
	} else {
		l.switchLedOn()
	}
}

// MonitorSoil monitors moisture for all sensors and stores it in the csv file.
func (g *Greenhouse) monitorSoil() {
	log.Println("Reading soil moisture...")
	if err := rpio.SpiBegin(rpio.Spi0); err != nil {
		panic(err)
	}
	rpio.SpiChipSelect(0) // Select CE0 slave
	buffer := make([]byte, 3)
	var result uint16
	mu.Lock()
	for _, s := range g.SoilSensors {
		for j := 0; j < 5; j++ {
			buffer[0] = 0x01
			buffer[1] = byte(8+s.Channel) << 4
			buffer[2] = 0x00
			rpio.SpiExchange(buffer) // buffer is populated with received data
			result = uint16((buffer[1]&0x3))<<8 + uint16(buffer[2])<<6
			appendCSV(moistFile, [][]string{{time.Now().Format("02-01-2006 15:04:05"), fmt.Sprintf("%v (%v)", s.Id, s.Channel), fmt.Sprint(j), fmt.Sprint(result)}})
			time.Sleep(time.Millisecond)
		}
		s.Time = time.Now()
		s.Value = int(result)
	}
	var values []int
	log.Println(len(g.SoilSensors))
	for _, s := range g.SoilSensors {
		values = append(values, s.Value)
		log.Printf("%v: %v", s.Id, s.Value)
	}
	g.SoilValue = calcAverage(values...)
	g.SoilTime = time.Now()
	mu.Unlock()
	rpio.SpiEnd(rpio.Spi0)
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

func seconds(d time.Duration) string {
	return fmt.Sprint(d.Seconds())
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

func readCSV(file string) [][]string {
	// Read the file
	f, err := os.Open(file)
	if err != nil {
		f, err := os.Create(file)
		if err != nil {
			log.Fatal("Unable to create csv", err)
		}
		f.Close()
		return [][]string{}
	}
	defer f.Close()
	r := csv.NewReader(f)
	lines, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	return lines
}

func appendCSV(file string, newLines [][]string) {

	// Get current data
	lines := readCSV(file)

	// Add new lines
	lines = append(lines, newLines...)

	// Write the file
	f, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	w := csv.NewWriter(f)
	if err = w.WriteAll(lines); err != nil {
		log.Fatal(err)
	}
}
