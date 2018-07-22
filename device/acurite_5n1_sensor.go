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
	// AcuRite5n1SensorName is the name that is used when storing into influxdb.
	AcuRite5n1SensorName = "AcuRite5n1Sensor"

	// AcuRite5n1SensorModelName is the model name rtl_433 returns.
	AcuRite5n1SensorModelName = "Acurite 5n1 sensor"
)

// AcuRite5n1SensorDataPoint represents a datapoint from an AcuRite5n1Sensor device.
type AcuRite5n1SensorDataPoint struct {
	Model                    string `json:"model"`
	TimeStr                  string `json:"time"`
	Time                     time.Time
	SensorID                 int     `json:"sensor_id"`
	Channel                  string  `json:"channel"`
	SequenceNum              int     `json:"sequence_num"`
	Battery                  string  `json:"battery"`
	MessageType              int     `json:"message_type"`
	WindSpeedMPH             float64 `json:"wind_speed_mph"`
	WindDirDeg               float64 `json:"wind_dir_deg"`
	WindDir                  string  `json:"wind_dir"`
	RainfallAccumulationInch float64 `json:"rainfall_accumulation_inch"`
	RaincounterRaw           int     `json:"raincounter_raw"`
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *AcuRite5n1SensorDataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the AcuRite5n1SensorDataPoint.
func (a *AcuRite5n1SensorDataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *AcuRite5n1SensorDataPoint) InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error) {
	tags := map[string]string{
		"model":        a.Model,
		"sensor_id":    strconv.Itoa(a.SensorID),
		"channel":      a.Channel,
		"sequence_num": strconv.Itoa(a.SequenceNum),
		"battery":      a.Battery,
		"message_type": strconv.Itoa(a.MessageType),
		"win_dir":      a.WindDir,
	}
	// Parsing any metadata for this type if we have some.
	for _, set := range sets {
		logger.Debug.Printf("processing metadata set %s", set)
		ProcessMetaDataFieldSet(tags, &set)
	}

	fields := map[string]interface{}{
		"wind_speed_mph":             a.WindSpeedMPH,
		"wind_dir_deg":               a.WindDirDeg,
		"rainfall_accumulation_inch": a.RainfallAccumulationInch,
		"raincounter_raw":            a.RaincounterRaw,
	}

	ParseTime(a)
	p, err := influx.NewPoint(AcuRite5n1SensorName, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
