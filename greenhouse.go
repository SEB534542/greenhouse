package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/SEB534542/seb"
	"github.com/stianeikeland/go-rpio"
)

// ghFile contains the json filename for storing the greenhouses and the components.
const ghFile = "greenhouse.json"

// configFile contains the json filename for storing the configuration of the program itself,
// such as tresholds, mail, etc.
const configFile = "config.json"

// configFolder is the folder where the config files (ghFile and configFile) are stored
const configFolder = "config"

// moistFile is the file where moisture data is stored
const moistFile = "soil_stats.csv"

// waterFile is the file where moisture data is stored
const wateringFile = "water.csv"

var mu sync.Mutex
var tpl *template.Template
var fm = template.FuncMap{"fdateHM": hourMinute, "fsec": seconds}
var g = &Greenhouse{}
var c = Config{}

type Config struct {
	RefreshRate time.Duration // Refresh rate for website
	Port        int
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
	SoilMin     int           // Minimal value for triggering
	SoilValue   int           // Last measured value
	SoilTime    time.Time     // Timing when last measured
	SoilFreq    time.Duration // Frequency for checking moisture
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
	err := seb.LoadConfig("./"+configFolder+"/"+configFile, &c)
	checkErr(err)
	if c.Port == 0 {
		c.Port = 8080
		log.Print("Unable to load port, set to %v", c.Port)
	}

	// Load greenhouse
	err = seb.LoadConfig("./"+configFolder+"/"+ghFile, &g)

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
	} else {
		log.Println("No LED configured")
	}

	// Monitor moisture
	if len(g.SoilSensors) != 0 {
		go func() {
			for {
				g.measureSoil()
				mu.Lock()
				log.Printf("Next soil measurement is in %v at %v", g.SoilFreq, g.SoilTime.Add(g.SoilFreq).Format("15:04"))
				for time.Until(g.SoilTime.Add(g.SoilFreq)) > 0 {
					mu.Unlock()
					time.Sleep(time.Second)
					mu.Lock()
				}
				mu.Unlock()
			}
		}()
	} else {
		log.Println("No SoilSensors configured")
	}

	log.Printf("Launching website at localhost:%v...", c.Port)
	http.HandleFunc("/", handlerMain)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.HandleFunc("/toggleled", handlerToggleLed)
	http.HandleFunc("/soilcheck", handlerSoilCheck)
	http.HandleFunc("/water", handlerWater)
	http.HandleFunc("/config/", handlerConfig)
	http.HandleFunc("/stop", handlerStop)
	log.Fatal(http.ListenAndServe(":"+fmt.Sprint(c.Port), nil))
}

func handlerMain(w http.ResponseWriter, req *http.Request) {
	mu.Lock()
	stats := seb.ReverseXss(seb.ReadCSV(moistFile))
	watering := seb.ReverseXss(seb.ReadCSV(wateringFile))
	data := struct {
		Time string
		Config
		*Greenhouse
		Stats    [][]string
		Watering [][]string
		NextSoil string
	}{
		time.Now().Format("_2 Jan 06 15:04:05"),
		c,
		g,
		stats,
		watering,
		g.SoilTime.Add(g.SoilFreq).Format("15:04"),
	}
	mu.Unlock()
	tpl.ExecuteTemplate(w, "index.gohtml", data)
}

func handlerToggleLed(w http.ResponseWriter, req *http.Request) {
	g.Led.toggleLed()
	http.Redirect(w, req, "/", http.StatusFound)
}

func handlerSoilCheck(w http.ResponseWriter, req *http.Request) {
	g.measureSoil()
	http.Redirect(w, req, "/", http.StatusFound)
}

func handlerConfig(w http.ResponseWriter, req *http.Request) {
	var err error
	var msgs []string
	mu.Lock()
	defer mu.Unlock()
	if req.Method == http.MethodPost {
		// saving config
		refreshRate, err := time.ParseDuration(req.PostFormValue("RefreshRate") + "s")
		if err != nil {
			msg := fmt.Sprintf("Unable to save RefreshRate '%v' (%v)", refreshRate, err)
			msgs = append(msgs, msg)
			log.Println(msg)
		} else {
			c.RefreshRate = refreshRate
		}
		port, err := seb.StrToIntZ(req.PostFormValue("Port"))
		if err != nil || !(port >= 1000 && port <= 9999) {
			msg := fmt.Sprintf("Unable to save port '%v', should be within range 1000-9999 (err)", port, err)
			msgs = append(msgs, msg)
			log.Println(msg)
		} else {
			c.Port = port
		}
		// Saving Greenhouse
		g.Id = req.PostFormValue("Id")
		g.Led.Id = req.PostFormValue("Led.Id")
		pin, err := seb.StrToIntZ(req.PostFormValue("Led.Pin"))
		if !(pin > 0 && pin < 28) || err != nil {
			msg := fmt.Sprintf("Unable to save Led Pin '%v' (%v)", pin, err)
			msgs = append(msgs, msg)
			log.Println(msg)
		} else {
			g.Led.Pin = rpio.Pin(pin)
		}
		g.Led.Start, err = seb.StoTime(req.PostFormValue("Led.Start"), 0)
		if err != nil {
			msg := fmt.Sprintf("Unable to save Led Start time '%v' (%v)", g.Led.Start, err)
			msgs = append(msgs, msg)
			log.Println(msg)
		}
		g.Led.End, err = seb.StoTime(req.PostFormValue("Led.End"), 0)
		if err != nil {
			msg := fmt.Sprintf("Unable to save Led End time %v (%v)", g.Led.End, err)
			msgs = append(msgs, msg)
			log.Println(msg)
		}
		for _, v := range g.SoilSensors {
			channel, err := seb.StrToIntZ(req.PostFormValue("SoilSensor." + v.Id + ".Channel"))
			if !(channel >= 0 && channel < 9) || err != nil {
				msg := fmt.Sprintf("Unable to save SoilSensor %v channel %v (%v)", v.Id, channel, err)
				msgs = append(msgs, msg)
				log.Println(msg)
			} else {
				v.Channel = channel
			}
			v.Id = req.PostFormValue("SoilSensor." + v.Id)
		}
		g.SoilMin, err = seb.StrToIntZ(req.PostFormValue("SoilMin"))
		if err != nil {
			msg := fmt.Sprintf("Unable to save Soil Threshold should be within range 1000-9999 (err)", err)
			msgs = append(msgs, msg)
			log.Println(msg)
		} else {
			c.Port = port
		}
		g.SoilFreq, err = time.ParseDuration(req.PostFormValue("SoilFreq") + "s")
		if err != nil {
			msg := fmt.Sprintf("Unable to save Soil Frequency: %v", err)
			msgs = append(msgs, msg)
			log.Println(msg)
		}
		seb.SaveToJSON(c, "./"+configFolder+"/"+configFile)
		seb.SaveToJSON(g, "./"+configFolder+"/"+ghFile)
		var msg string
		if len(msgs) == 0 {
			msg = "Saved configuration"
		} else {
			msg = "Saved the rest"
		}
		msgs = append(msgs, msg)
		log.Println(msg)
	}
	data := struct {
		Msgs []string
		Config
		*Greenhouse
	}{
		msgs,
		c,
		g,
	}
	err = tpl.ExecuteTemplate(w, "config.gohtml", data)
	if err != nil {
		log.Panic(err)
	}
	return
}

func handlerWater(w http.ResponseWriter, req *http.Request) {
	xs := []string{fmt.Sprint(time.Now().Format("02-01-2006 15:04:05")), "Water added"}
	seb.AppendCSV(wateringFile, [][]string{xs})
	http.Redirect(w, req, "/", http.StatusFound)
}

func handlerStop(w http.ResponseWriter, req *http.Request) {
	g.Led.switchLedOff()
	rpio.Close()
	log.Println("Shutting down...")
	os.Exit(3)
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
func (g *Greenhouse) measureSoil() {
	log.Println("Reading soil moisture...")
	if err := rpio.SpiBegin(rpio.Spi0); err != nil {
		panic(err)
	}
	rpio.SpiChipSelect(0) // Select CE0 slave
	buffer := make([]byte, 3)
	var values []int
	mu.Lock()
	for _, s := range g.SoilSensors {
		var results []int
		for j := 0; j < 5; j++ {
			buffer[0] = 0x01
			buffer[1] = byte(8+s.Channel) << 4
			buffer[2] = 0x00
			rpio.SpiExchange(buffer) // buffer is populated with received data
			result := uint16((buffer[1]&0x3))<<8 + uint16(buffer[2])<<6
			results = append(results, int(result))
			time.Sleep(time.Millisecond)
		}
		s.Time = time.Now()
		s.Value = seb.CalcAverage(results...)
		values = append(values, s.Value)
	}
	g.SoilValue = seb.CalcAverage(values...)
	g.SoilTime = time.Now()
	log.Printf("Average soil is %v", g.SoilValue)
	xs := []string{fmt.Sprint(g.SoilTime.Format("02-01-2006 15:04:05")), fmt.Sprint(g.SoilValue)}
	for _, v := range values {
		xs = append(xs, fmt.Sprint(v))
	}
	seb.AppendCSV(moistFile, [][]string{xs})
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

// HourMinute returns a variable of type time.Time as a string in format "15:04".
// This function is used for displaying time on a gohtml webpage.
func hourMinute(t time.Time) string {
	return t.Format("15:04")
}

// Seconds returns a variable of type time.Duration as a string in seconds."
// This function is used for displaying durations on a gohtml webpage.
func seconds(d time.Duration) string {
	return fmt.Sprint(d.Seconds())
}
