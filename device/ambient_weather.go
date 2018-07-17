package device

import (
	"fmt"
	"strconv"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
)

var (
	// AmbientWeatherName is the name that is used when storing into influxdb.
	AmbientWeatherName = "AmbientWeather"
)

// AmbientWeatherDataPoint represents a datapoint from an AmbientWeather device.
type AmbientWeatherDataPoint struct {
	Model        string  `json:"model"`
	TimeStr      string  `json:"time"`
	RTL433ID     int     `json:"rtl_433_id"`
	Device       int     `json:"device"`
	Channel      int     `json:"channel"`
	Battery      string  `json:"battery"`
	TemperatureF float64 `json:"temperature_f"`
	Humidity     int     `json:"humidity"`
	Time         time.Time
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *AmbientWeatherDataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the AmbientWeatherDataPoint.
func (a *AmbientWeatherDataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *AmbientWeatherDataPoint) InfluxData() (*influx.Point, error) {
	tags := map[string]string{
		"model":      a.Model,
		"rtl_433_id": strconv.Itoa(a.RTL433ID),
		"device":     strconv.Itoa(a.Device),
		"channel":    strconv.Itoa(a.Channel),
		"battery":    a.Battery,
	}

	fields := map[string]interface{}{
		"temperature_f": a.TemperatureF,
		"humidity":      a.Humidity,
	}
	ParseTime(a)
	p, err := influx.NewPoint(AmbientWeatherName, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
