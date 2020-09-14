package main

import (
	"fmt"
	"time"

	"github.com/SEB534542/seb"
)

// TODO: change all pins to actual RPIO pins

// TODO: create interface for all sensors(?)

// A led represents the a LED light in the greenhouse
type led struct {
	// Name specifies the identifier (name or a number) of the led
	id     string // e.g. "Main" or "01"
	active bool
	pin    int
	start  time.Time
	end    time.Time
}

// A pump represents the waterpump that can be activated through the pin to
// add water to the greenhouse
type pump struct {
	id  string
	pin int
}

type moistSensor struct {
	id    string
	value float64
	pin   int
}

// A servo represents a servo motor to open a window for ventilation
type servo struct {
	id  string
	pin int
}

type tempSensor struct {
	id    string
	value float64
	pin   int
}

type greenhouse struct {
	leds    []led
	pumps   []pump
	moistSs []moistSensor
	servos  []servo
	tempSs  []tempSensor
}

func main() {
	g1 := &greenhouse{
		leds: []led{
			led{
				id:     "Led 1",
				pin:    1,
				start:  time.Now().Add(-60 * time.Minute),
				end:    time.Now().Add(60 * time.Minute),
				active: false,
			},
		},
		pumps: []pump{
			pump{
				id:  "Pump 1",
				pin: 2,
			},
		},
		moistSs: []moistSensor{
			moistSensor{
				id:    "Moisture sensor 1",
				pin:   3,
				value: 0,
			},
		},
		servos: []servo{
			servo{
				id:  "Servo pump 1",
				pin: 4,
			},
		},
		tempSs: []tempSensor{
			tempSensor{
				id:    "Temp sensor 1",
				pin:   5,
				value: 0,
			},
		},
	}
	fmt.Println(g1)
	fmt.Println(g1.leds[0].start.Format("Mon Jan 2 15:04 MST"))
	fmt.Println(g1.leds[0].end.Format("Mon Jan 2 15:04 MST"))
	err := seb.SaveToJSON("test", "test.json") //g1, "greenhouse.json")
	if err != nil {
		fmt.Println("Error:", err)
	}

}
