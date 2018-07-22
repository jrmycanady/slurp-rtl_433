package device

import (
	"fmt"
	"strconv"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/jrmycanady/slurp-rtl_433/config"
	"github.com/jrmycanady/slurp-rtl_433/logger"
)

var (
	// CalibeurRF104Name is the name that is used when storing into influxdb.
	CalibeurRF104Name = "CalibeurRF104"

	// CalibeurRF104ModelName is the model name rtl_433 returns.
	CalibeurRF104ModelName = "Calibeur RF-104"
)

// CalibeurRF104DataPoint represents a datapoint from an CalibeurRF104 device.
type CalibeurRF104DataPoint struct {
	Model        string `json:"model"`
	TimeStr      string `json:"time"`
	Time         time.Time
	ID           int     `json:"id"`
	TemperatureC float64 `json:"temperature_C"`
	Humidity     int     `json:"humidity"`
	MIC          string  `json:"mic"`
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *CalibeurRF104DataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the CalibeurRF104DataPoint.
func (a *CalibeurRF104DataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *CalibeurRF104DataPoint) InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error) {
	tags := map[string]string{
		"model": a.Model,
		"id":    strconv.Itoa(a.ID),
	}
	// Parsing any metadata for this type if we have some.
	for _, set := range sets {
		logger.Debug.Printf("processing metadata set %s", set)
		ProcessMetaDataFieldSet(tags, &set)
	}

	fields := map[string]interface{}{
		"temperature_C": a.TemperatureC,
		"humidity":      a.Humidity,
	}

	ParseTime(a)
	p, err := influx.NewPoint(CalibeurRF104Name, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
