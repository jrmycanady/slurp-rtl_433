package main

import (
	"time"
)

// Add interface for Device that allows GetTime and SetTime so we can parse time on all types.

// AmbientWeather represents a data output from the AmbientWeather device.
type AmbientWeather struct {
	TimeStr      string  `json:"time"`
	RTL433ID     int     `json:"rtl_433_id"`
	Device       int     `json:"device"`
	Channel      int     `json:"channel"`
	Battery      string  `json:"battery"`
	TemperatureF float64 `json:"temperature_f"`
	Humidity     int     `json:"humidity"`
	Time         time.Time
}

func (a *AmbientWeather) Parse() {
	var err error
	a.Time, err = time.Parse("2006-01-02 15:04:05", a.TimeStr)

	// a.Time, err = time.Parse(strings.Replace(a.TimeStr, " ", "T"), 1)
	if err != nil {
		panic(err)
	}
}
