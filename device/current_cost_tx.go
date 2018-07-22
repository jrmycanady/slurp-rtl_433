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
	// CurrentCostTXName is the name that is used when storing into influxdb.
	CurrentCostTXName = "CurrentCostTX"

	// CurrentCostTXModelName is the model name rtl_433 returns.
	CurrentCostTXModelName = "CurrentCost TX"
)

// CurrentCostTXDataPoint represents a datapoint from an CurrentCostTX device.
type CurrentCostTXDataPoint struct {
	Model   string `json:"model"`
	TimeStr string `json:"time"`
	Time    time.Time
	DevID   int `json:"dev_id"`
	Power0  int `json:"power0"`
	Power1  int `json:"power1"`
	Power2  int `json:"power2"`
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *CurrentCostTXDataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the CurrentCostTXDataPoint.
func (a *CurrentCostTXDataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *CurrentCostTXDataPoint) InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error) {
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
		"power0": a.Power0,
		"power1": a.Power1,
		"power2": a.Power2,
	}

	ParseTime(a)
	p, err := influx.NewPoint(CurrentCostTXName, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
