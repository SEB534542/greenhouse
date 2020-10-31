package main

import (
	"fmt"

	"github.com/stianeikeland/go-rpio"
)

var pins = []rpio.Pin{rpio.Pin(23)} // rpio.Pin(24)}

func main() {
	relays()
}

// relays toggles the specified pins
func relays() {
	rpio.Open()
	defer rpio.Close()

	for _, p := range pins {
		fmt.Println("Toggling pin", p)
		p.Output()
		p.Toggle()
		//p.Low() // Aan
		//p.High() // Uit
	}
}
