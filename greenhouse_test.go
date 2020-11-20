package main

import (
	"fmt"
	"log"
	"net/http"
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
		MoistSs: []*MoistSensor{
			{Id: "Left", Channel: 2},
			{Id: "Middle", Channel: 1},
			{Id: "Right", Channel: 0},
		},
	}
	seb.SaveToJSON(&g1, fname)
	os.Remove(fname)
}

//func TestReadState(t *testing.T) {
//	ps := 23
//	rpio.Open()
//	defer rpio.Close()
//	fmt.Println("Pin", ps, "state is", rpio.Pin(ps).Read())
//	if rpio.Pin(ps).Read() == rpio.Low {
//		// Low = 0 means it it is on, so turning it off by setting it to high
//		rpio.Pin(ps).Write(rpio.High)
//		time.Sleep(time.Second)
//		rpio.Pin(ps).Toggle()
//	} else if rpio.Pin(ps).Read() == rpio.High {
//		// High = 1 means it it is off, so turning it on by setting it to low
//		rpio.Pin(ps).Write(rpio.Low)
//		time.Sleep(time.Second)
//		rpio.Pin(ps).Toggle()
//	}
//}

//func TestMonitorLed(t *testing.T) {
//	rpio.Open()
//	defer rpio.Close()
//	l := &Led{
//		Id:    "Main Led",
//		Pin:   rpio.Pin(23),
//		Start: time.Now().Add(time.Second * 2),
//		End:   time.Now().Add(time.Second * 4),
//	}
//	state := l.Pin.Read()
//	l.monitorLed()
//	go l.monitorLed()
//	time.Sleep(time.Second * 3)
//	l.Pin.Write(state)
//}

//func TestMoistFile(t *testing.T) {
//	channel := 0
//	i := 1
//	var v uint16 = 15674
//	appendCSV(moistFile, [][]string{
//		{fmt.Sprint(channel), time.Now().Format("02-01-2006 15:04:05"), fmt.Sprint(i), fmt.Sprint(v)}})
//}

//func TestGetMoist(t *testing.T) {
//	rpio.Open()
//	defer rpio.Close()
//	g := &Greenhouse{
//		MoistSs: []*MoistSensor{
//			{Id: "Left", Channel: 2},
//			{Id: "Middle", Channel: 1},
//			{Id: "Right", Channel: 0},
//		}}
//	g.monitorMoist()
//}

func TestHandlerMain(t *testing.T) {
	c.RefreshRate = 30
	log.Println("Launching website...")
	http.HandleFunc("/", handlerMain)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	// http.HandleFunc("/stop", handlerStop)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
