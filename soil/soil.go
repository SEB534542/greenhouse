package soil

import (
	"fmt"
	"sync"
	"time"

	"github.com/stianeikeland/go-rpio"
)

var mu sync.Mutex

// Sensors represents a soil sensors that is connected to
// the MCP3008 chip.
type Sensor struct {
	Id    string    // Identifier for Sensor, e.g. "left" or "right"
	Chan  int       // Sensor Channel is connected to on the MCP3008 chip
	Value int       // Last measured value
	Time  time.Time // Time when value was measured
}

// Update updates the Value and Time for each Sensor and returns the average
// value and time of measurement.
func Update(xs []*Sensor) (int, time.Time) {
	fmt.Println("Reading soil moisture...")
	if err := rpio.SpiBegin(rpio.Spi0); err != nil {
		panic(err)
	}
	rpio.SpiChipSelect(0) // Select CE0 slave
	buffer := make([]byte, 3)
	var values []int
	for _, s := range xs {
		var results []int
		for j := 0; j < 5; j++ {
			buffer[0] = 0x01
			buffer[1] = byte(8+s.Chan) << 4
			buffer[2] = 0x00
			rpio.SpiExchange(buffer) // buffer is populated with received data
			result := uint16((buffer[1]&0x3))<<8 + uint16(buffer[2])<<6
			results = append(results, int(result))
			time.Sleep(time.Millisecond)
		}
		mu.Lock()
		s.Time = time.Now()
		s.Value = calcAverage(results...)
		mu.Unlock()
		fmt.Printf("Sensor %v has value %v\n", s.Id, s.Value)
		values = append(values, s.Value)
	}
	rpio.SpiEnd(rpio.Spi0)
	return calcAverage(values...), time.Now()
}

// calcAverage takes a variadic parameter of integers and
// returns the average integer.
func calcAverage(xi ...int) int {
	total := 0
	for _, v := range xi {
		total = total + v
	}
	return total / len(xi)
}
