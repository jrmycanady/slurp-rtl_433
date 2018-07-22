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
	// EfergyE2CTName is the name that is used when storing into influxdb.
	EfergyE2CTName = "EfergyE2CT"

	// EfergyE2CTModelName is the model name rtl_433 returns.
	EfergyE2CTModelName = "Efergy e2 CT"
)

// EfergyE2CTDataPoint represents a datapoint from an EfergyE2CT device.
type EfergyE2CTDataPoint struct {
	Model   string `json:"model"`
	TimeStr string `json:"time"`
	Time    time.Time
	ID      int    `json:"id"`
	Current int    `json:"current"`
	Battery string `json:"battery"`
	Learn   string `json:"learn"`
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *EfergyE2CTDataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the EfergyE2CTDataPoint.
func (a *EfergyE2CTDataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *EfergyE2CTDataPoint) InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error) {
	tags := map[string]string{
		"model":   a.Model,
		"id":      strconv.Itoa(a.ID),
		"battery": a.Battery,
		"learn":   a.Learn,
	}
	// Parsing any metadata for this type if we have some.
	for _, set := range sets {
		logger.Debug.Printf("processing metadata set %s", set)
		ProcessMetaDataFieldSet(tags, &set)
	}

	fields := map[string]interface{}{
		"current": a.Current,
		"learn":   a.Learn,
	}

	ParseTime(a)
	p, err := influx.NewPoint(EfergyE2CTName, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
