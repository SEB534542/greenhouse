package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/stianeikeland/go-rpio"
)

func main() {
	//jdev2()
	stian()
}

func stian() {
	var channel int
	if len(os.Args) < 2 {
		channel = 1
		fmt.Println("No input, accessing channel", channel)
	} else {
		arg := os.Args[1]
		channel, err := strconv.Atoi(arg)
		if err != nil {
			channel = 0
			fmt.Println("Wrong input, accessing channel", channel)
		}
	}
	if err := rpio.Open(); err != nil {
		panic(err)
	}
	if err := rpio.SpiBegin(rpio.Spi0); err != nil {
		panic(err)
	}
	rpio.SpiChipSelect(0) // Select CE0 slave

	buffer := make([]byte, 3)

	for i := 0; i < 20; i++ {
		buffer[0] = 0x01
		buffer[1] = byte(8+channel) << 4
		buffer[2] = 0x00

		rpio.SpiExchange(buffer) // buffer is populated with received data

		result := uint16((buffer[1]&0x3))<<8 + uint16(buffer[2])<<6
		fmt.Printf("%v\t%v\n", i+1, result)
		time.Sleep(time.Second)
	}
	rpio.SpiEnd(rpio.Spi0)
	rpio.Close()
}
