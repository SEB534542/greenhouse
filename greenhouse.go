package main

import (
	"fmt"
	"log"
	"time"

	"github.com/SEB534542/seb"
)

// TODO: change all pins to actual RPIO pins

// TODO: create interface for all sensors(?)

// TODO: create a slice of greenhouses

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
	Id  string
	Pin int
}

type MoistSensor struct {
	Id    string
	Value float64
	Pin   int
}

// A servo represents a servo motor to open a window for ventilation
type Servo struct {
	Id  string
	Pin int
}

type TempSensor struct {
	Id    string
	Value float64
	Pin   int
}

type Greenhouse struct {
	Leds     []Led
	Pumps    []Pump
	MoistSs  []MoistSensor
	Servos   []Servo
	TempSs   []TempSensor
	MoistMin float64
	TempMin  float64
	TempMax  float64
}

func main() {
	g1 := &Greenhouse{
		Leds: []Led{
			Led{
				Id:     "Led 1",
				Pin:    1,
				Start:  time.Now().Add(-60 * time.Minute),
				End:    time.Now().Add(60 * time.Minute),
				Active: false,
			},
		},
		Pumps: []Pump{
			Pump{
				Id:  "Pump 1",
				Pin: 2,
			},
		},
		MoistSs: []MoistSensor{
			MoistSensor{
				Id:    "Moisture sensor 1",
				Pin:   3,
				Value: 0,
			},
		},
		Servos: []Servo{
			Servo{
				Id:  "Servo pump 1",
				Pin: 4,
			},
		},
		TempSs: []TempSensor{
			TempSensor{
				Id:    "Temp sensor 1",
				Pin:   5,
				Value: 0,
			},
		},
		MoistMin: 15,
		TempMin:  15,
		TempMax:  20,
	}
	fmt.Println(g1)
	// Save g1 to JSON
	checkErr(seb.SaveToJSON(g1, "greenhouses.json"))
}

func checkErr(err error) {
	if err != nil {
		log.Println("Error:", err)
	}
	return
}
