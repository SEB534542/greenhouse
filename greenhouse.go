package main

import (
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/SEB534542/seb"
)

// TODO: change all pins to actual RPIO pins

// ghFile contains the json filename for storing the greenhouse config.
const ghFile = "greenhouses.json"
const configFile = "config.json"

var mu sync.Mutex

var config = struct {
	MoistCheck time.Duration
	TempCheck  time.Duration
}{}

// A led represents the a LED light in the greenhouse
type Led struct {
	// Name specifies the identifier (name or a number) of the led
	Id     string // e.g. "Main" or "01"
	Active bool
	Pin    int
	Start  time.Time
	End    time.Time
}

// A pump represents the waterpump that can be activated through the Pin to
// add water to the greenhouse.
type Pump struct {
	Id  string `json:"PumpId"`
	Pin int
	Dur time.Duration
}

type MoistSensor struct {
	Id    string
	Value int
	Pin   int
}

// A servo represents a servo motor to open a window for ventilation.
type Servo struct {
	Id   string
	Pin  int
	Open bool
}

// A TempSensor represents a sensor that measures the temperature.
type TempSensor struct {
	Id    string
	Value int
	Pin   int
}

// A Box represents a greenbox with plants in a greenhouse,
// with it's own water pump and moisture sensors.
type Box struct {
	Id         string
	MoistSs    []*MoistSensor
	MoistMin   int
	MoistValue int
	Pump
}

// A Greenhouse represents a greenhouse consisting of a/multiple box(es)
// with plants, sensors and LED lights.
type Greenhouse struct {
	Id        string
	Leds      []*Led
	Servos    []*Servo
	TempSs    []*TempSensor
	Boxes     []*Box
	TempMin   int
	TempMax   int
	TempValue int
}

func main() {
	log.Println("--------Start of program--------")

	// Load general config
	data, err := ioutil.ReadFile("./config/" + configFile)
	checkErr(err)
	checkErr(json.Unmarshal(data, &config))

	// Loading greenhouse config
	gx := []*Greenhouse{}
	data, err = ioutil.ReadFile("./config/" + ghFile)
	checkErr(err)
	checkErr(json.Unmarshal(data, &gx))

	for _, g := range gx {
		//Resetting Start and End date to today for each LED
		for _, l := range g.Leds {
			l.Start = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), l.Start.Hour(), l.Start.Minute(), 0, 0, time.Now().Location())
			l.End = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), l.End.Hour(), l.End.Minute(), 0, 0, time.Now().Location())
			log.Println("Reset dates to today:", l.Start, l.End)
		}
		log.Println(len(g.Boxes), "box(es) configured")

		// Monitor moisture for each box
		for _, b := range g.Boxes {
			go b.monitorMoist()
		}

		// Monitor moisture for each LED
		for _, l := range g.Leds {
			go l.monitorLed()
		}

		// Monitor temperature for each sensor
		go g.monitorTemp()
	}
	log.Println("Start eternal loop")
	for {
	}
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
		g.TempValue = seb.CalcAverage(values...)
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
		defer mu.Unlock()
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

// MonitorMoist monitors moisture and if insufficent enables the waterpump
func (b *Box) monitorMoist() {
	for {
		values := []int{}
		mu.Lock()
		for _, s := range b.MoistSs {
			s.getMoist()
			values = append(values, s.Value)
		}
		b.MoistValue = seb.CalcAverage(values...)
		log.Printf("Average moisture in box %v: %v based on: %v", b.Id, b.MoistValue, values)
		if b.MoistValue <= b.MoistMin {
			mu.Unlock()
			// TODO: start pump for t seconds
			mu.Lock()
			log.Printf("Pump %s started for %s in Box %s\n", b.Pump.Id, b.Pump.Dur, b.Id)
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
// and stores it in MoistSensor.Value
func (s *MoistSensor) getMoist() {
	// TODO: get Moist value from RPIO
	// Seed the random number generator using the current time (nanoseconds since epoch)
	rand.Seed(time.Now().UnixNano())

	// Much harder to predict...but it is still possible if you know the day, and hour, minute...
	s.Value = rand.Intn(1000)
}

// CheckErr evaluates err for errors (not nil)
// and triggers a log.Panic containing the error
func checkErr(err error) {
	if err != nil {
		log.Panic("Error:", err)
	}
	return
}
