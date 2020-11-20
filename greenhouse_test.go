package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/SEB534542/seb"
	//"github.com/stianeikeland/go-rpio"
)

var _ = fmt.Printf // For debugging; delete when done.
var _ os.File      // For debugging; delete when done.

func TestloadConfig(t *testing.T) {
	err := loadConfig("./"+configFolder+"/"+configFile, &config)
	if err != nil {
		log.Printf("Error while loading ", err)
	}

	// Load greenhouse
	err = loadConfig("./"+configFolder+"/"+ghFile, &gx)
	if err != nil {
		log.Printf("Error while loading ", err)
	}

}

// func TestEmptyGreenhouse(t *testing.T) {
// 	const fname1 = "./config/test_greenhouses.json"

// 	g1 := []*Greenhouse{
// 		{
// 			Id: "Main Greenhouse",
// 			Boxes: []*Box{
// 				{
// 					Id: "Box1",
// 					MoistSs: []*MoistSensor{
// 						{
// 							Id:      "Moisture sensor 1",
// 							Channel: 1,
// 							Value:   0,
// 						},
// 					},
// 					Leds: []*Led{
// 						{
// 							Id:     "Led 1",
// 							Pin:    rpio.Pin(23),
// 							Start:  time.Date(2020, time.November, 8, 0, 0, 0, 0, time.Local),
// 							End:    time.Date(2020, time.November, 22, 0, 0, 0, 0, time.Local),
// 							Active: false,
// 						},
// 					},
// 					MoistMin: 1000,
// 				},
// 			},
// 		},
// 	}
// 	// Save g1 to JSON
// 	checkErr(seb.SaveToJSON(g1, fname1))

// 	// Delete files
// 	os.Remove(fname1)
// }
