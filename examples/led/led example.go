package main

import (
	"fmt"
	"time"

	"github.com/SEB534542/greenhouse/led"
	"github.com/stianeikeland/go-rpio"
)

func main() {
	// Connecting to rpio Pins
	rpio.Open()
	defer rpio.Close()

	l := &led.Led{
		Id:  "Main",
		Pin: rpio.Pin(23),
	}

	fmt.Println("Toggling light")
	l.Toggle()
	time.Sleep(time.Second)
	l.Off()
	time.Sleep(time.Second)
	l.On()
	time.Sleep(time.Second)
	l.Off()
}
