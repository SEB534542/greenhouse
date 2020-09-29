package main

import (
	"fmt"

	"github.com/stianeikeland/go-rpio/v4"
)

func main() {
	relays()
}

func relays() {
	rpio.Open()
	defer rpio.Close()

	r1 := rpio.Pin(23)
	r2 := rpio.Pin(24)
	r3 := rpio.Pin(27)
	r4 := rpio.Pin(17)

	pins := []rpio.Pin{r1, r2, r3, r4}
	for _, p := range pins {
		fmt.Println("Toggling pin", p)
		p.Output()
		p.Toggle()
		// p.Low()  // Aan
		// p.High() // Uit
	}
}
