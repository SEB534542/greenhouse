package main

import (
	"fmt"
	"os"
	"testing"
	//"github.com/SEB534542/seb"
	//"github.com/stianeikeland/go-rpio"
)

var _ = fmt.Printf // For debugging; delete when done.
var _ os.File      // For debugging; delete when done.

func TestloadConfig(t *testing.T) {
	err := loadConfig("./"+configFolder+"/"+configFile, &config)
	if err != nil {
		fmt.Printf("Error while loading ", err)
	}

	// Load greenhouse
	err = loadConfig("./"+configFolder+"/"+ghFile, &gx)
	if err != nil {
		fmt.Printf("Error while loading ", err)
	}

}
