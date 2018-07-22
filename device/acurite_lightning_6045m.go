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
	// AcuRiteLightning6045MName is the name that is used when storing into influxdb.
	AcuRiteLightning6045MName = "AcuRiteLightning6045M"

	// AcuRiteLightning6045MModelName is the model name rtl_433 returns.
	AcuRiteLightning6045MModelName = "Acurite Lightning 6045M"
)

// AcuRiteLightning6045MDataPoint represents a datapoint from an AcuRiteLightning6045M device.
type AcuRiteLightning6045MDataPoint struct {
	Model        string `json:"model"`
	TimeStr      string `json:"time"`
	Time         time.Time
	ID           int    `json:"id"`
	Channel      string `json:"channel"`
	TemperatureF int64  `json:"temperature_F"`
	Humidity     int    `json:"humidity"`
	StrikeCount  int    `json:"strike_count"`
	StormDist    int    `json:"storm_dist"`
	ActiveMode   int    `json:"active"`
	RFI          int    `json:"rfi"`
	USSB1        int    `json:"ussb1"`
	Battery      string `json:"battery"`
	Exception    int    `json:"exception"`
}

// GetTimeStr returns the string format of the time as provided by the device
// output.
func (a *AcuRiteLightning6045MDataPoint) GetTimeStr() string {
	return a.TimeStr
}

// SetTime sets the time value fo the AcuRiteLightning6045MDataPoint.
func (a *AcuRiteLightning6045MDataPoint) SetTime(t time.Time) {
	a.Time = t
}

// InfluxData builds a new InfluxDB datapoint from the values in the DataPoint.
func (a *AcuRiteLightning6045MDataPoint) InfluxData(sets map[string]config.MetaDataFieldSet) (*influx.Point, error) {
	tags := map[string]string{
		"model":       a.Model,
		"id":          strconv.Itoa(a.ID),
		"channel":     a.Channel,
		"active_mode": strconv.Itoa(a.ActiveMode),
		"rfi":         strconv.Itoa(a.RFI),
		"ussb1":       strconv.Itoa(a.USSB1),
		"battery":     a.Battery,
		"exception":   strconv.Itoa(a.Exception),
	}
	// Parsing any metadata for this type if we have some.
	for _, set := range sets {
		logger.Debug.Printf("processing metadata set %s", set)
		ProcessMetaDataFieldSet(tags, &set)
	}

	fields := map[string]interface{}{
		"temperature_F": a.TemperatureF,
		"humidity":      a.Humidity,
		"strike_count":  a.StrikeCount,
		"storm_disk":    a.StormDist,
	}

	ParseTime(a)
	p, err := influx.NewPoint(AcuRiteLightning6045MName, tags, fields, a.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to create point: %s", err)
	}

	return p, nil
}
