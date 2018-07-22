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
	// AcuRite609TXCSensorName is the name that is used when storing into influxdb.
	AcuRite609TXCSensorName = "AcuRite609TXCSensor"

	// AcuRite609TXCSensorModelName is the model name rtl_433 returns.
	AcuRite609TXCSensorModelName = "Acurite 609TXC Sensor"
)

// AcuRite609TXCSensorDataPoint represents a datapoint from an AcuRite609TXCSensor device.
type AcuRite609TXCSensorDataPoint struct {
	Model        string `json:"model"`
	TimeStr      string `json:"time"`
	Time         time.Time
	ID           int     `json:"id"`
	Battery      string  `json:"battery"`
	Status       int     `json:"status"`
	TemperatureC float64 `json:"temperature_c"`
	Humidity     int     `json:"humidity"`
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *AcuRite609TXCSensorDataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the AcuRite609TXCSensorDataPoint.
func (a *AcuRite609TXCSensorDataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *AcuRite609TXCSensorDataPoint) InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error) {
	tags := map[string]string{
		"model":   a.Model,
		"id":      strconv.Itoa(a.ID),
		"status":  strconv.Itoa(a.Status),
		"battery": a.Battery,
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
	p, err := influx.NewPoint(AcuRite609TXCSensorName, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
