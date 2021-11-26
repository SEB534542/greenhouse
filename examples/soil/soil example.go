package main

import (
	"fmt"

	"github.com/SEB534542/greenhouse/soil"
	"github.com/stianeikeland/go-rpio"
)

func main() {
	// Connecting to rpio Pins
	rpio.Open()
	defer rpio.Close()

	soilSensors := []*soil.Sensor{
		&soil.Sensor{Id: "Left", Chan: 2},
		&soil.Sensor{Id: "Middle", Chan: 1},
		&soil.Sensor{Id: "Right", Chan: 0},
	}

	v, t := soil.Update(soilSensors)
	fmt.Printf("Soil value = %v at time = %v\n", v, t.Format("02-01-2006 15:04:05"))
}
