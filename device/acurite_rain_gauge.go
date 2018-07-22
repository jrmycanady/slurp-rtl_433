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
	// AcuRiteRainGaugeName is the name that is used when storing into influxdb.
	AcuRiteRainGaugeName = "AcuRiteRainGauge"

	// AcuRiteRainGaugeModelName is the model name rtl_433 returns.
	AcuRiteRainGaugeModelName = "Acurite Rain Gauge"
)

// AcuRiteRainGaugeDataPoint represents a datapoint from an AcurRiteRainGauge device.
type AcuRiteRainGaugeDataPoint struct {
	Model   string `json:"model"`
	TimeStr string `json:"time"`
	Time    time.Time
	Rain    float64 `json:"rain"`
	ID      int     `json:"id"`
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *AcuRiteRainGaugeDataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the AcuRiteRainGaugeDataPoint.
func (a *AcuRiteRainGaugeDataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *AcuRiteRainGaugeDataPoint) InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error) {
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
		"rain_mm": a.Rain,
	}

	ParseTime(a)
	p, err := influx.NewPoint(AcuRiteRainGaugeName, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
