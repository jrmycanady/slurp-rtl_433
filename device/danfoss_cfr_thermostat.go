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
	// DanfossCFRThermostatName is the name that is used when storing into influxdb.
	DanfossCFRThermostatName = "DanfossCFRThermostat"

	// DanfossCFRThermostatModelName is the model name rtl_433 returns.
	DanfossCFRThermostatModelName = "Danfoss CFR Thermostat"
)

// DanfossCFRThermostatDataPoint represents a datapoint from an DanfossCFRThermostat device.
type DanfossCFRThermostatDataPoint struct {
	Model        string `json:"model"`
	TimeStr      string `json:"time"`
	Time         time.Time
	ID           int     `json:"id"`
	TemperatureC float64 `json:"temperature_c"`
	SetPointC    float64 `json:"setpoint_C"`
	Switch       string  `json:"string"`
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *DanfossCFRThermostatDataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the DanfossCFRThermostatDataPoint.
func (a *DanfossCFRThermostatDataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *DanfossCFRThermostatDataPoint) InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error) {
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
		"temperature_C": a.TemperatureC,
		"setpoint_C":    a.SetPointC,
	}

	ParseTime(a)
	p, err := influx.NewPoint(DanfossCFRThermostatName, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
