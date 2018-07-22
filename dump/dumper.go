package dump

import (
	"fmt"
	"sync"
	"time"

	influxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/jrmycanady/slurp-rtl_433/config"
	"github.com/jrmycanady/slurp-rtl_433/device"
	"github.com/jrmycanady/slurp-rtl_433/logger"
)

// Dumper represents the process that flushes datapoints to the differnet
// output such as InfluxDB.
type Dumper struct {
	dataPointsChan <-chan device.DataPoint
	cancelChan     chan struct{}
	doneChan       chan struct{}
	cfg            config.Config
	iClient        influxClient.Client
	lock           *sync.Mutex
	running        bool
	bp             influxClient.BatchPoints
}

// NewDumper creates a new dumper instance that is ready to start.
func NewDumper(cfg config.Config, dataPointChan <-chan device.DataPoint) *Dumper {
	return &Dumper{
		dataPointsChan: dataPointChan,
		cancelChan:     make(chan struct{}),
		doneChan:       make(chan struct{}),
		cfg:            cfg,
		lock:           &sync.Mutex{},
	}
}

// SetRunning sets the running state of the dumper.
func (d *Dumper) SetRunning(state bool) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.running = state
}

// StartDump attempts to start the dumber. An error is returned if it failed
// to do so.
func (d *Dumper) StartDump() error {

	// Building influxdb client.
	iClient, err := buildInfluxClient(d.cfg)
	if err != nil {
		return err
	}
	d.iClient = iClient

	// Starting dumper process.
	go d.dump()

	return nil
}

// StopDump requests the dumper to stop.
func (d *Dumper) StopDump() {
	d.cancelChan <- struct{}{}
}

// dumper listens on the dataPointsChan and proceses the datapoints as they come
// in. Dumper will flush the points once it reaches one of two situations. First
// the dumper has up to FlushDataPointCount or it has been FlushTimeTrigger from
// the last timer flush. It's currently possible for the time flush to run
// shortly after a flush from a maximum data points.
func (d *Dumper) dump() {
	var err error
	d.SetRunning(true)
	defer d.SetRunning(false)

	logger.Info.Println("dumper has entered the running state")

	d.bp, err = influxClient.NewBatchPoints(influxClient.BatchPointsConfig{
		Database:  d.cfg.InfluxDB.Database,
		Precision: "s",
	})
	if err != nil {
		panic(err)
	}
	lastFlushTime := time.Now()
	flushTicker := time.NewTicker(time.Duration(10) * time.Second)
	for {
		select {
		case <-flushTicker.C:
			logger.Debug.Println("flush ticker ticked")
			if len(d.bp.Points()) == 0 {
				continue
			}
			if time.Since(lastFlushTime).Seconds() >= d.cfg.InfluxDB.FlushTimeTrigger {
				if !d.flushUntilCancel() {
					logger.Info.Println("dumper received a request to cancel during dumper flush")
					return
				}
				lastFlushTime = time.Now()
			}
		case dp := <-d.dataPointsChan:
			logger.Debug.Println("new datapoint received")

			switch v := dp.(type) {
			case *device.AmbientWeatherDataPoint:
				p, err := v.InfluxData(d.cfg.Meta[device.AmbientWeatherModelName])
				if err != nil {
					continue
				}
				d.bp.AddPoint(p)
			}

			logger.Debug.Printf("time until time flush: %f/%f", time.Since(lastFlushTime).Seconds(), d.cfg.InfluxDB.FlushTimeTrigger)

			// Flush if full or not sent in a while.
			if len(d.bp.Points()) >= d.cfg.InfluxDB.FlushDataPointCount || time.Since(lastFlushTime).Seconds() >= d.cfg.InfluxDB.FlushTimeTrigger {

				if !d.flushUntilCancel() {
					logger.Info.Println("dumper received a request to cancel during dumper flush")
					return
				}

				lastFlushTime = time.Now()
			}
		case <-d.cancelChan:
			logger.Info.Println("dumper has received a request to cancel")

			// attempt to flush any points in flight.
			if err = d.flush(); err != nil {
				//TODO save points in flight.
			}
			return

		}
	}
}

// flush flushes the datapoints to influx if possible.
func (d *Dumper) flush() error {
	var err error

	if err := d.iClient.Write(d.bp); err != nil {
		return fmt.Errorf("failed to send points to InfluxDB %s", err)
	}

	// clearing out points.
	d.bp, err = influxClient.NewBatchPoints(influxClient.BatchPointsConfig{
		Database:  d.cfg.InfluxDB.Database,
		Precision: "s",
	})
	if err != nil {
		panic(err)
	}

	logger.Info.Printf("dumped %d datapoints to InfluxDB", len(d.bp.Points()))
	return nil
}

// flushUntilCancel flushes all the points found in the dumper. It only returns
// once a flush is successful or a cancel is received. If ok the return
// value will be false.
// Upon failure it will wait 1 additional second for every 10 failures. It will
// max out at 30 seconds of wait.
func (d *Dumper) flushUntilCancel() (ok bool) {
	var err error
	err = fmt.Errorf("error")
	failures := 0
	failureWaitTime := 1
	for err != nil {
		if err = d.flush(); err != nil {
			failures++

			if failures%10 == 0 {
				if failures < 30 {
					failureWaitTime++
				}
			}

			logger.Error.Printf("failed to send data to InfluxDB: %s", err)
			logger.Info.Printf("waiting %d second before retry", failureWaitTime)
			time.Sleep(time.Duration(failureWaitTime) * time.Second)
		}
		// Check for a cancel before trying again.
		select {
		case <-d.cancelChan:
			//TODO save to disk points in flight.
			return false
		default:
		}
	}

	return true
}

// buildInfluxClient generates a new InfluxDB client based on the configuration provided.
func buildInfluxClient(config config.Config) (influxClient.Client, error) {
	var err error
	address := ""
	if config.InfluxDB.HTTPS {
		address = fmt.Sprintf("https://%s:%d", config.InfluxDB.FQDN, config.InfluxDB.Port)
	} else {
		address = fmt.Sprintf("http://%s:%d", config.InfluxDB.FQDN, config.InfluxDB.Port)
	}
	client, err := influxClient.NewHTTPClient(influxClient.HTTPConfig{
		Addr:     address,
		Username: config.InfluxDB.Username,
		Password: config.InfluxDB.Password,
	})
	if err != nil {
		return nil, err
	}

	_, _, err = client.Ping(time.Duration(10) * time.Second)
	if err != nil {
		return nil, err
	}

	return client, nil

}
