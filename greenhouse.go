package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"
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
	// Load greenhouse config
	g1 := &Greenhouse{}
	data, err := ioutil.ReadFile("./config/" + ghFile)
	checkErr(err)
	checkErr(json.Unmarshal(data, g1))

	fmt.Println(g1)
}

func MonitorMoist

func checkErr(err error) {
	if err != nil {
		log.Panic("Error:", err)
	}
	return
}
