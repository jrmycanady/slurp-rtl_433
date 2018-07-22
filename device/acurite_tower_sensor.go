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
	// AcuRiteTowerSensorName is the name that is used when storing into influxdb.
	AcuRiteTowerSensorName = "AcuRiteTowerSensor"

	// AcuRiteTowerSensorModelName is the model name rtl_433 returns.
	AcuRiteTowerSensorModelName = "Acurite tower sensor"
)

// AcuRiteTowerSensorDataPoint represents a datapoint from an AcuRiteTowerSensor device.
type AcuRiteTowerSensorDataPoint struct {
	Model        string `json:"model"`
	TimeStr      string `json:"time"`
	Time         time.Time
	ID           int     `json:"id"`
	SensorID     int     `json:"sensor_id"`
	Channel      string  `json:"channel"`
	TemperatureC float64 `json:"temperature_C"`
	Humidity     int     `json:"humidity"`
	BatteryLow   int     `json:"battery_low"`
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *AcuRiteTowerSensorDataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the AcuRiteTowerSensorDataPoint.
func (a *AcuRiteTowerSensorDataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *AcuRiteTowerSensorDataPoint) InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error) {
	tags := map[string]string{
		"model":       a.Model,
		"id":          strconv.Itoa(a.ID),
		"sensor_id":   strconv.Itoa(a.SensorID),
		"channel":     a.Channel,
		"battery_low": strconv.Itoa(a.BatteryLow),
	}
	// Parsing any metadata for this type if we have some.
	for _, set := range sets {
		logger.Debug.Printf("processing metadata set %s", set)
		ProcessMetaDataFieldSet(tags, &set)
	}

	fields := map[string]interface{}{
		"temperature_C": a.TemperatureC,
		"humidity":      a.Humidity,
	}

	ParseTime(a)
	p, err := influx.NewPoint(AcuRiteTowerSensorName, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
