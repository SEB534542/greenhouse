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

func TestLoadConfigEmpty(t *testing.T) {

	type test struct {
		fname string
		i     interface{}
	}

	tests := []test{
		{"./" + configFolder + "/configtest.json", &c},
		{"./" + configFolder + "/greenhousetest.json", &g},
	}

	// Load empty
	for _, t := range tests {
		err := loadConfig(t.fname, t.i)

		if err != nil {
			t.Errorf("File %v did not exist, or no error occured", t.fname)
		}
	}

	// Load again to check if files still exist and remove files
	for _, t := range tests {
		err := loadConfig(t.fname, t.i)
		if err != nil {
			fmt.Println("Error while loading:", err)
		}
		os.Remove(t.fname)
	}
}
