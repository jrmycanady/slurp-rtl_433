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
	// Bresser3CHSensorName is the name that is used when storing into influxdb.
	Bresser3CHSensorName = "Bresser3CHSensor"

	// Bresser3CHSensorModelName is the model name rtl_433 returns.
	Bresser3CHSensorModelName = "Bresser 3CH sensor"
)

// Bresser3CHSensorDataPoint represents a datapoint from an Bresser3CHSensor device.
type Bresser3CHSensorDataPoint struct {
	Model        string `json:"model"`
	TimeStr      string `json:"time"`
	Time         time.Time
	ID           int     `json:"id"`
	Channel      string  `json:"channel"`
	Battery      string  `json:"battery"`
	TemperatureF float64 `json:"temperature_F"`
	Humidity     int     `json:"humidity"`
	MIC          string  `json:"mic"`
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *Bresser3CHSensorDataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the Bresser3CHSensorDataPoint.
func (a *Bresser3CHSensorDataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *Bresser3CHSensorDataPoint) InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error) {
	tags := map[string]string{
		"model":   a.Model,
		"id":      strconv.Itoa(a.ID),
		"channel": a.Channel,
		"battery": a.Battery,
	}
	// Parsing any metadata for this type if we have some.
	for _, set := range sets {
		logger.Debug.Printf("processing metadata set %s", set)
		ProcessMetaDataFieldSet(tags, &set)
	}

	fields := map[string]interface{}{
		"temperature_F": a.TemperatureF,
		"humidity":      a.Humidity,
	}

	ParseTime(a)
	p, err := influx.NewPoint(Bresser3CHSensorName, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
