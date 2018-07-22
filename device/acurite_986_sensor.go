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
	// AcuRite986SensorName is the name that is used when storing into influxdb.
	AcuRite986SensorName = "AcuRite986Sensor"

	// AcuRite986SensorModelName is the model name rtl_433 returns.
	AcuRite986SensorModelName = "Acurite 986 Sensor"
)

// AcuRite986SensorDataPoint represents a datapoint from an AcuRite986Sensor device.
type AcuRite986SensorDataPoint struct {
	Model        string `json:"model"`
	TimeStr      string `json:"time"`
	Time         time.Time
	ID           int     `json:"id"`
	Channel      string  `json:"channel"`
	TemperatureF float64 `json:"temperature_F"`
	Battery      string  `json:"battery"`
	Status       int     `json:"status"`
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *AcuRite986SensorDataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the AcuRite986SensorDataPoint.
func (a *AcuRite986SensorDataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *AcuRite986SensorDataPoint) InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error) {
	tags := map[string]string{
		"model":   a.Model,
		"id":      strconv.Itoa(a.ID),
		"channel": a.Channel,
		"status":  strconv.Itoa(a.Status),
		"battery": a.Battery,
	}
	// Parsing any metadata for this type if we have some.
	for _, set := range sets {
		logger.Debug.Printf("processing metadata set %s", set)
		ProcessMetaDataFieldSet(tags, &set)
	}

	fields := map[string]interface{}{
		"temperature_F": a.TemperatureF,
	}

	ParseTime(a)
	p, err := influx.NewPoint(AcuRite986SensorName, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
