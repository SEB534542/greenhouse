package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/SEB534542/seb"
)

func TestMain(t *testing.T) {
	const fname = "greenhouse.json"
	g1 := &Greenhouse{
		Leds: []*Led{
			{
				Id:     "Led 1",
				Pin:    1,
				Start:  time.Now().Add(-60 * time.Minute),
				End:    time.Now().Add(60 * time.Minute),
				Active: false,
			},
		},
		Servos: []*Servo{
			{
				Id:  "Servo pump 1",
				Pin: 2,
			},
		},
		TempSs: []*TempSensor{
			{
				Id:    "Temp sensor 1",
				Pin:   3,
				Value: 0,
			},
		},
		Boxes: []*Box{
			{
				Id: "Tomatoes",
				Pump: Pump{
					Id:  "Pump 1",
					Pin: 4,
					Dur: 5 * time.Second,
				},
				MoistSs: []*MoistSensor{
					{
						Id:    "Moisture sensor 1",
						Pin:   5,
						Value: 0,
					},
					{
						Id:    "Moisture sensor 2",
						Pin:   6,
						Value: 0,
					},
				},
				MoistMin: 1000,
			},
		},
		TempMin: 15,
		TempMax: 20,
	}
	fmt.Println(g1)
	// Save g1 to JSON
	checkErr(seb.SaveToJSON(g1, "./config/"+fname))
}
