package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/SEB534542/seb"
)

func TestMain(t *testing.T) {
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
	checkErr(seb.SaveToJSON(g1, "./config/"+"testfile.json"))
}
