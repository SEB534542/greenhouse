package led

import (
	"sync"
	"time"

	"github.com/stianeikeland/go-rpio"
)

var mu sync.Mutex

// Led represents a LED light
type Led struct {
	Id     string    // Identifier for the LED e.g. "Main" or "01"
	Active bool      // Indicator if LED is turned on (true) or not (false)
	Pin    rpio.Pin  // Pin LED is connected to
	Start  time.Time // Start time for LED to be active
	End    time.Time // End time for LED to be no longer active
}

// SwitchLedOn switches LED on.
func (l *Led) On() {
	l.Pin.Output()
	l.Pin.Write(rpio.Low)
	l.Active = true
	return
}

// SwitchLedOn switches LED off.
func (l *Led) Off() {
	l.Pin.Output()
	l.Pin.Write(rpio.High)
	l.Active = false
	return
}

func (l *Led) Toggle() {
	if l.Active == true {
		l.Off()
	} else {
		l.On()
	}
}

// Monitor checks if LED should be switched on or off based.
func (l *Led) Monitor() {
	mu.Lock()
	switch {
	case time.Now().After(l.End):
		fmt.Println("Resetting Start and End to tomorrow for LED", l.Id)
		l.Start = l.Start.AddDate(0, 0, 1)
		l.End = l.End.AddDate(0, 0, 1)
		fallthrough
	case time.Now().Before(l.Start):
		fmt.Printf("Turning LED %s off and snoozing for %v until %s...\n", l.Id, time.Until(l.Start), l.Start.Format("02-01 15:04"))
		l.Off()
		for time.Until(l.Start) > 0 {
			mu.Unlock()
			time.Sleep(time.Second)
			mu.Lock()
		}
		fallthrough
	case time.Now().After(l.Start) && time.Now().Before(l.End):
		fmt.Printf("Turning LED %s on and snoozing for %v until %s...\n", l.Id, time.Until(l.End), l.End.Format("02-01 15:04"))
		l.On()
		for time.Until(l.End) > 0 {
			mu.Unlock()
			time.Sleep(time.Second)
			mu.Lock()
		}
	}
	mu.Unlock()
}
