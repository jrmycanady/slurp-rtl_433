// Package devices contains the definitions of all devices that rtl_433 can
// monitor. Each device must provide the InfluxDatapoint method.
package device

import (
	"encoding/json"
	"fmt"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/jrmycanady/slurp-rtl_433/config"
	"github.com/jrmycanady/slurp-rtl_433/logger"
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
	InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error)
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
	case AmbientWeatherModelName:
		a := AmbientWeatherDataPoint{}
		if err = json.Unmarshal(d, &a); err != nil {
			return nil, err
		}
		return &a, nil
	}

	return nil, fmt.Errorf("unknown model: %s", b.Model)
}

// ProcessMetaDataFieldSet processes the field set by adding the tags
// if the comparison values are true.
func ProcessMetaDataFieldSet(pTags map[string]string, f *config.MetaDataFieldSet) {
	fmt.Println("Starting metadat compare")
	fmt.Println(pTags)
	fmt.Println(f.CompEqualTags)

	// Processing each tag that needs a equalCompare.
	for mT := range f.CompEqualTags {
		// Looking at each tag to see if we have one to match.
		for pT := range pTags {
			// Comparing names to determine if they match. If they don't then
			// the fieldset doesn't apply and we return.
			logger.Debug.Printf("%s ?= %s", pT, mT)
			if pT == mT {
				logger.Debug.Printf("[%s][%s] ?= [%s][%s]", pT, pTags[pT], mT, f.CompEqualTags[mT])

				if pTags[pT] != f.CompEqualTags[mT] {

					return
				}
				logger.Debug.Printf("[%s][%s] == [%s][%s]", pT, pTags[pT], mT, f.CompEqualTags[mT])
			}
		}
	}

	// If we made it to here then it does apply so add all the tags.
	for k := range f.Tags {
		pTags[k] = f.Tags[k]
	}

}
