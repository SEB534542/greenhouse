package main

import (
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"github.com/SEB534542/seb"
)

// TODO: change all pins to actual RPIO pins

// TODO: create interface for all sensors(?)

// TODO: create a slice of greenhouses

// TODO: decide on pointers or using pointer[index]

// ghFile contains the json filename for storing the greenhouse config
const ghFile = "greenhouse.json"

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
	TempMin float64
	TempMax float64
}

func main() {
	log.Println("--------Start of program--------")

	// Loading greenhouse config
	g := &Greenhouse{}
	data, err := ioutil.ReadFile("./config/" + ghFile)
	checkErr(err)
	checkErr(json.Unmarshal(data, g))

	//Resetting LED Start and End to today
	for i, _ := range g.Leds {
		g.Leds[i].Start = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), g.Leds[i].Start.Hour(), g.Leds[i].Start.Minute(), 0, 0, time.Now().Location())
		g.Leds[i].End = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), g.Leds[i].End.Hour(), g.Leds[i].End.Minute(), 0, 0, time.Now().Location())
		log.Println(g.Leds[i].Start, g.Leds[i].End)
	}
	log.Println(g)
	log.Println(len(g.Boxes), "box(es) configured")
	for _, b := range g.Boxes {
		b.monitorMoist()
	}
	g.monitorLed()
	log.Println(g.Leds[0].Start)
}

// MonitorLED checks if LED should be enabled or disabled
func (g *Greenhouse) monitorLed() {
	for i, _ := range g.Leds {
		switch {
		case time.Now().After(g.Leds[i].End):
			g.Leds[i].switchLedOff()
			// Resetting Start and End to tomorrow
			g.Leds[i].Start = g.Leds[i].Start.AddDate(0, 0, 1)
			g.Leds[i].End = g.Leds[i].End.AddDate(0, 0, 1)
		}

	}
}

func (l *Led) switchLedOn() {
	if !l.Active {
		l.switchLed()
	}
}

func (l *Led) switchLedOff() {
	if l.Active {
		l.switchLed()
	}
}

func (l *Led) switchLed() {
	if l.Active {
		// TODO: turn LED off
		log.Printf("Turning Led %s off", l.Id)
		l.Active = false
	} else {
		// TODO: turn LED on
		log.Printf("Turning Led %s on", l.Id)
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
