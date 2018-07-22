package device

import (
	"fmt"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/jrmycanady/slurp-rtl_433/config"
	"github.com/jrmycanady/slurp-rtl_433/logger"
)

var (
	// EfergyOpticalName is the name that is used when storing into influxdb.
	EfergyOpticalName = "EfergyOptical"

	// EfergyOpticalModelName is the model name rtl_433 returns.
	EfergyOpticalModelName = "Efergy e2 CT"
)

// EfergyOpticalDataPoint represents a datapoint from an EfergyOptical device.
type EfergyOpticalDataPoint struct {
	Model   string `json:"model"`
	TimeStr string `json:"time"`
	Time    time.Time
	Pulses  int     `json:"pulses"`
	Energy  float64 `json:"energy"`
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *EfergyOpticalDataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the EfergyOpticalDataPoint.
func (a *EfergyOpticalDataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *EfergyOpticalDataPoint) InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error) {
	tags := map[string]string{
		"model": a.Model,
	}
	// Parsing any metadata for this type if we have some.
	for _, set := range sets {
		logger.Debug.Printf("processing metadata set %s", set)
		ProcessMetaDataFieldSet(tags, &set)
	}

	fields := map[string]interface{}{
		"pulses": a.Pulses,
		"energy": a.Energy,
	}

	ParseTime(a)
	p, err := influx.NewPoint(EfergyOpticalName, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
