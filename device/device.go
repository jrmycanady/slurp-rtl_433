// Package devices contains the definitions of all devices that rtl_433 can
// monitor. Each device must provide the InfluxDatapoint method.
package device

import (
	"encoding/json"
	"fmt"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
)

// DataPoint is in interface for interacting with differnet types of devices.
//
// InfluxData creates a new influx data point containing the values for the
// device.
//
// GetTimeStr returns the string representation of the time from the DataPoint.
//
// SetTime sets the time property of the DataPoint.
type DataPoint interface {
	InfluxData() (*influx.Point, error)
	GetTimeStr() string
	SetTime(t time.Time)
}

// BaseDataPoint is the minimum properties a BaseDataPoint must implement. It is
// primarily used to parse json output to determine the real device.
type BaseDataPoint struct {
	Model string `json:"model"`
}

// ParseTime parses the string time of the DataPoint then stores it in the
// time property
func ParseTime(d DataPoint) {
	t, err := time.Parse("2006-01-02 15:04:05", d.GetTimeStr())
	if err != nil {
		panic(err)
	}
	d.SetTime(t)
}

// ParseDataPoint parses the string into the proper DataPoint type. If parsing
// fails nil will be returned with an error.
func ParseDataPoint(d []byte) (DataPoint, error) {
	var err error

	// Marshalling the data point into a base data point se the type can be
	//determined.
	b := BaseDataPoint{}
	if err = json.Unmarshal([]byte(d), &b); err != nil {
		return nil, err
	}

	switch b.Model {
	case "Ambient Weather F007TH Thermo-Hygrometer":
		a := AmbientWeatherDataPoint{}
		if err = json.Unmarshal(d, &a); err != nil {
			return nil, err
		}
		return &a, nil
	}

	return nil, fmt.Errorf("unknown model: %s", b.Model)
}
