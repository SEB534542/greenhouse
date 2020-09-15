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

// TODO: create interface for all sensors(?)

// TODO: create a slice of greenhouses

// TODO: create go and locks

// ghFile contains the json filename for storing the greenhouse config
const ghFile = "greenhouse.json"

//var mu sync.Mutex

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
// add water to the greenhouse
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

// A servo represents a servo motor to open a window for ventilation
type Servo struct {
	Id  string
	Pin int
}

type TempSensor struct {
	Id    string
	Value int
	Pin   int
}

type Box struct {
	Id string
	Pump
	MoistSs  []*MoistSensor
	MoistMin int
}

type Greenhouse struct {
	Leds    []*Led
	Servos  []*Servo
	TempSs  []*TempSensor
	Boxes   []*Box
	TempMin int
	TempMax int
}

func main() {
	log.Println("--------Start of program--------")

	// Loading greenhouse config
	g := &Greenhouse{}
	data, err := ioutil.ReadFile("./config/" + ghFile)
	checkErr(err)
	checkErr(json.Unmarshal(data, g))

	//Resetting Start and End date to today for each LED
	for _, l := range g.Leds {
		l.Start = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), l.Start.Hour(), l.Start.Minute(), 0, 0, time.Now().Location())
		l.End = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), l.End.Hour(), l.End.Minute(), 0, 0, time.Now().Location())
		log.Println(l.Start, l.End)
	}
	log.Println(g)
	log.Println(len(g.Boxes), "box(es) configured")

	// Monitor moisture for each box
	for _, b := range g.Boxes {
		b.monitorMoist()
	}

	// Monitor moisture for each LED
	for _, l := range g.Leds {
		l.monitorLed()
	}
	log.Println(g.Leds[0].Start)

	// Monitor temperature for each sensor
	g.monitorTemp()
}

func (g *Greenhouse) monitorTemp() {
	//values := []int
}

// MonitorLED checks if LED should be enabled or disabled
func (l *Led) monitorLed() {
	for {
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
				time.Sleep(time.Second)
			}
			fallthrough
		case time.Now().After(l.Start) && time.Now().Before(l.End):
			log.Printf("Turning LED %s on for %s sec until %s", l.Id, time.Until(l.End).Seconds(), l.End.Format("02-01 15:04"))
			l.switchLedOn()
			for i := 0; i < int(time.Until(l.End).Seconds()); i++ {
				time.Sleep(time.Second)
			}
		}
	}
}

func (l *Led) switchLedOn() {
	if !l.Active {
		l.switchLed()
	}
	return
}

func (l *Led) switchLedOff() {
	if l.Active {
		l.switchLed()
	}
	return
}

func (l *Led) switchLed() {
	if l.Active {
		// TODO: turn LED off
		log.Printf("Turning Led %s off...", l.Id)
		l.Active = false
	} else {
		// TODO: turn LED on
		log.Printf("Turning Led %s on...", l.Id)
		l.Active = true
	}
}

// MonitorMoist monitors moisture and if insufficent enables the pump
func (b *Box) monitorMoist() {
	values := []int{}
	for _, s := range b.MoistSs {
		s.getMoist()
		values = append(values, s.Value)
	}
	moist := seb.CalcAverage(values...)
	log.Printf("Average moisture in box %v: %v based on: %v", b.Id, moist, values)
	if moist <= b.MoistMin {
		// TODO: start pump for t seconds
		log.Printf("Pump %s started for %s in Box %s\n", b.Pump.Id, b.Pump.Dur, b.Id)
	}
	// TODO: add sleep to next day (variable)
}

func (s *MoistSensor) getMoist() {
	// TODO: get Moist value from RPIO
	// Seed the random number generator using the current time (nanoseconds since epoch)
	rand.Seed(time.Now().UnixNano())

	// Much harder to predict...but it is still possible if you know the day, and hour, minute...
	s.Value = rand.Intn(1000)
}

func checkErr(err error) {
	if err != nil {
		log.Panic("Error:", err)
	}
	return
}
