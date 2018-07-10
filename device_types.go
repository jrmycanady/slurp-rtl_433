package main

import (
	"time"
)

// Device is an interface for Devices that allows GetTime and SetTime so we can parse time on all types.
type Device interface {
	GetTimeStr() string
	SetTime(t time.Time)
}

type BaseDevice struct {
	Device int `json:"device"`
}

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

// GetTimeStr retrieves the string value of the time provided by rtl_433
func (a *AmbientWeather) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time property with the time t.
func (a *AmbientWeather) SetTime(t time.Time) {
	a.Time = t
}

// ParseTime pares the string time of a device and sets it's time value with the result.
func ParseTime(d Device) {
	t, err := time.Parse("2006-01-02 15:04:05", d.GetTimeStr())
	if err != nil {
		panic(err)
	}
	d.SetTime(t)
}
