package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/SEB534542/seb"
	"github.com/stianeikeland/go-rpio"
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
	for _, x := range tests {
		err := loadConfig(x.fname, x.i)
		if err != nil {
			t.Error("Error", err)
		}
	}

	// Load again to check if files still exist and remove files
	for _, x := range tests {
		err := loadConfig(x.fname, x.i)
		if err != nil {
			t.Error("Error", err)
		}
		os.Remove(x.fname)
	}
}

func TestSaveConfig(t *testing.T) {
	fname := "./" + configFolder + "/greenhouse_test.json"

	g1 := Greenhouse{
		Id: "My Greenhouse",
		Led: &Led{
			Id:    "Main Led",
			Pin:   rpio.Pin(23),
			Start: time.Date(2020, time.November, 20, 8, 30, 0, 0, time.Local),
			End:   time.Date(2020, time.November, 20, 21, 45, 0, 0, time.Local),
		},
	}
	seb.SaveToJSON(&g1, fname)
	os.Remove(fname)
}

func TestReadState(t *testing.T) {
	rpio.Open()
	defer rpio.Close()
	fmt.Println(rpio.Pin(23).Read())
}
