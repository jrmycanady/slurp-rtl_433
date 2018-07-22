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
	// Akhan100F14Name is the name that is used when storing into influxdb.
	Akhan100F14Name = "Akhan100F14"

	// Akhan100F14ModelName is the model name rtl_433 returns.
	Akhan100F14ModelName = "Akhan 100F14 remote keyless entry"
)

// Akhan100F14DataPoint represents a datapoint from an Akhan100F14 device.
type Akhan100F14DataPoint struct {
	Model   string `json:"model"`
	TimeStr string `json:"time"`
	Time    time.Time
	ID      int    `json:"id"`
	Data    string `json:"string"`
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *Akhan100F14DataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the Akhan100F14DataPoint.
func (a *Akhan100F14DataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *Akhan100F14DataPoint) InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error) {
	tags := map[string]string{
		"model": a.Model,
		"id":    strconv.Itoa(a.ID),
		"data":  a.Data,
	}
	// Parsing any metadata for this type if we have some.
	for _, set := range sets {
		logger.Debug.Printf("processing metadata set %s", set)
		ProcessMetaDataFieldSet(tags, &set)
	}

	fields := map[string]interface{}{}

	ParseTime(a)
	p, err := influx.NewPoint(Akhan100F14Name, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
