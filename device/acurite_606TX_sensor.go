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
	// AcuRite606TXSensorName is the name that is used when storing into influxdb.
	AcuRite606TXSensorName = "AcuRite606TXSensor"

	// AcuRite606TXSensorModelName is the model name rtl_433 returns.
	AcuRite606TXSensorModelName = "Acurite 606TX Sensor"
)

// AcuRite606TXSensorDataPoint represents a datapoint from an AcuRite606TXSensor device.
type AcuRite606TXSensorDataPoint struct {
	Model        string `json:"model"`
	TimeStr      string `json:"time"`
	Time         time.Time
	ID           int     `json:"id"`
	Battery      string  `json:"battery"`
	TemperatureC float64 `json:"temperature_c"`
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *AcuRite606TXSensorDataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the AcuRite606TXSensorDataPoint.
func (a *AcuRite606TXSensorDataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *AcuRite606TXSensorDataPoint) InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error) {
	tags := map[string]string{
		"model":   a.Model,
		"id":      strconv.Itoa(a.ID),
		"battery": a.Battery,
	}
	// Parsing any metadata for this type if we have some.
	for _, set := range sets {
		logger.Debug.Printf("processing metadata set %s", set)
		ProcessMetaDataFieldSet(tags, &set)
	}

	fields := map[string]interface{}{
		"temperature_C": a.TemperatureC,
	}

	ParseTime(a)
	p, err := influx.NewPoint(AcuRite606TXSensorName, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
