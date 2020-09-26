package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/SEB534542/seb"
)

var _ = fmt.Printf // For debugging; delete when done.

func TestMain(t *testing.T) {
	const fname1 = "./config/test_greenhouses.json"
	const fname2 = "./config/test_config_test.json"

	g1 := []*Greenhouse{
		{
			Id: "Main Greenhouse",
			Servos: []*Servo{
				{
					Id:   "Servo pump 1",
					Pin:  2,
					Open: false,
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
					Leds: []*Led{
						{
							Id:     "Led 1",
							Pin:    1,
							Start:  time.Now().Add(-60 * time.Minute),
							End:    time.Now().Add(60 * time.Minute),
							Active: false,
						},
					},
					MoistMin: 1000,
				},
			},
			TempMin: 15,
			TempMax: 25,
		},
	}
	// Save g1 to JSON
	checkErr(seb.SaveToJSON(g1, fname1))

	// Create config file
	config.MoistCheck = time.Second * 12
	config.TempCheck = time.Second * 10
	config.RefreshRate = time.Second * 10
	checkErr(seb.SaveToJSON(config, fname2))

	// Delete files
	os.Remove(fname1)
	os.Remove(fname2)
}
