package main

import (
	"fmt"
	"machine"

	_ "tinygo.org/x/drivers/mcp3008"
)

func main() {
	fmt.Println("-----Start of program-----")

	bus := machine.SPI{}
	fmt.Printf("%v - %T\n", bus, bus)
	config := machine.SPIConfig{
		SCK: machine.Pin(11),
		SDO: machine.Pin(9),
		SDI: machine.Pin(10),
	}
	fmt.Printf("%v - %T\n", config, config)

	//	bus.Configure(config)

	//	err := bus.Configure(config)
	//	if err != nil {
	//		fmt.Println("Error configuring SPI:", err)
	//	}
	//	mcp := mcp3008.New(bus, machine.Pin(5))
	//	mcp.Configure()
	//	fmt.Println("MCP configured:", mcp)
	//	value, err := mcp.Read(0)
	//	if err != nil {
	//		fmt.Println("Error reading pin:", err)
	//	}
	//	fmt.Println("Value of Pin is:", value)
}
