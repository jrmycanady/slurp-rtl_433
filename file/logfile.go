package file

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jrmycanady/slurp-rtl_433/device"
	"github.com/jrmycanady/slurp-rtl_433/logger"
)

const (
	// The byte value for a linefeed.
	lf byte = 10

	// The byte value for a carriage return.
	cr byte = 13
)

// A LogFile represents a rtl_433 json output file. Various actions can be
// performed against it such as slurping.
type LogFile struct {
	// Inode is the inode of the file system for the log file itself.
	Inode uint64 `json:"inode"`

	// Offset is the last read location that was successfully processed.
	Offset int64 `json:"offset"`

	// MetaDataID is the id of the meta data file.
	MetaDataID uuid.UUID `json:"metaDataId"`

	// MetaDataFilePath is the full file path of the MetData file.
	MetaDataFilePath string `json:"metaDataFilePath"`

	// LogFilePath is the last known path to the log file.
	LogFilePath string `json:"logFilePath"`

	lock *sync.Mutex

	// slurpCancelChan provides a channel that can be closed to tell the slurp
	// process to stop.
	slurpCancelChan chan struct{}

	// slurpRunning denotes if a slurp is currently running on the file.
	slurpRunning bool

	// found is true when the filer has found the file, thus making it
	// available for slurping.
	found bool

	// SlurperShutdownMaxWaitSeconds is the maximum seconds a slurper will wait to shutdown.
	SlurperShutdownMaxWaitSeconds float64
}

// Found returns that found status of the LogFile. True if the LogFile has
// beenf found.
func (l *LogFile) Found() bool {
	return l.found
}

// SetFound sets the found status of the log file.
func (l *LogFile) SetFound(f bool) {
	l.lock.Lock()
	l.found = f
	l.lock.Unlock()
}

// NewLogFile creates a new log file. Optionally a marshalled json string
// of the meta data can be provided which will be the source of the
// configuration. If it's empty a new LogFile will be generated.
func NewLogFile(metaDataJSON []byte) (*LogFile, error) {
	l := LogFile{
		lock:            &sync.Mutex{},
		slurpCancelChan: make(chan struct{}),
	}

	if len(metaDataJSON) != 0 {
		if err := json.Unmarshal([]byte(metaDataJSON), &l); err != nil {
			return nil, fmt.Errorf("failed to parse metaDataJSON: %s", err)
		}
	} else {
		uuid, err := uuid.NewRandom()
		if err != nil {
			panic(err)
		}

		l.MetaDataID = uuid
	}

	return &l, nil
}

// Save saves the log file meta data.
func (l *LogFile) Save() error {
	l.lock.Lock()
	defer l.lock.Unlock()
	j, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(l.MetaDataFilePath, j, 0644)
}

// SlurpRunning returns true if a slurp is currently running on the file.
func (l *LogFile) SlurpRunning() bool {
	return l.slurpRunning
}

// setSlurpRunning set the current running state of the slurp action.
func (l *LogFile) setSlurpRunning(state bool) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.slurpRunning = state
}

// slurp monitors the log file for data and sends the results to the
// dataPointChan. It will continue to slurp until StopSlurp is called.
//
// Slurp may be called multiple times but will only ever start one slurp
// task.
// sleepTimeSeconds is the amount of time the slurper sleeps before looking for more
// lines to process.
func (l *LogFile) slurp(dataPointChan chan<- device.DataPoint, sleepTimeSeconds int) {
	l.setSlurpRunning(true)
	defer l.setSlurpRunning(false)

	// Preventing crazy low sleep time.
	if sleepTimeSeconds < 1 {
		sleepTimeSeconds = 1
	}

	// Opening the file for processing.
	f, err := os.Open(l.LogFilePath)
	if err != nil {
		logger.Error.Printf("failed to open file %s", l.LogFilePath)
		logger.Debug.Printf("failed to open file %s: %s", l.LogFilePath, err)
		return
	}
	defer f.Close()

	// Validating it's still the same file.
	stat, err := f.Stat()
	if err != nil {
		logger.Error.Printf("failed to stat file %s", l.LogFilePath)
		return
	}
	sys := stat.Sys().(*syscall.Stat_t)
	if sys.Ino != l.Inode {
		logger.Error.Printf("the inode has changed so the file is new %s", l.LogFilePath)
		return
	}

	logger.Info.Printf("opened and starting slurping of %s", l.LogFilePath)

	// Processing the file until slurpCancelChan is closed.
	for {
		var n int
		var buff = make([]byte, 10)
		var line = make([]byte, 0, 200)

		// Check to see if we should stop.
		select {
		case <-l.slurpCancelChan:
			logger.Verbose.Printf("stop for slurper on %s received", l.LogFilePath)
			return
		default:
		}

		// Seeking to the last location not recorded.
		_, err = f.Seek(l.Offset, 0)
		if err != nil {
			logger.Error.Printf("failed to seek to file %s", l.LogFilePath)
			logger.Debug.Printf("failed to seek to file %s: %s", l.LogFilePath, err)
			return
		}
		logger.Debug.Printf("seeking complete on %s", l.LogFilePath)

		// Reading until we reach the end of the file.
		for err != io.EOF {
			// Check to see if we should stop.
			select {
			case <-l.slurpCancelChan:
				logger.Verbose.Printf("stop for slurper on %s received", l.LogFilePath)
				return
			default:
			}

			// Reading up to the buffer length.
			n, err = f.Read(buff)

			// Check each character in the buffer for line feed \n or carriage return \r.
			// Finding it means the line has ended and we should save it off. Then continue on.
			var startIndex = 0
			for i := 0; i < n; i++ {

				// [v][f][f][cr][  lf ][x][x][x] n = 8
				// [    0:3    ][ 3:4 ][  i+1: ]  0:9
				// [    0:i    ][i:i+1]
				switch {
				case buff[i] == cr:
					// Add anything before the \r to the line and saving it.
					line = append(line, buff[startIndex:i]...)
					if len(line) > 0 {
						if err := savePoint(line, dataPointChan); err != nil {
							logger.Error.Printf("failed to save data point: %s", err)
						}
						l.Save()
					}
					line = line[:0]

					// Checking to see if the returned data is larger enough for another character
					// and if so contains a line feed. If so kick it out by pushing i up one.
					if (i+1 < n) && buff[i+1] == lf {
						i++
					}

					// Update the start index to be the next value.
					startIndex = i + 1

					continue

				case buff[i] == lf:
					// Should only reach here if no \r was found. So simply add everything before to the line,
					// save it. Then add anything left to he new line.
					line = append(line, buff[startIndex:i]...)
					// Saving the line if not empty.
					if len(line) > 0 {
						if err := savePoint(line, dataPointChan); err != nil {
							logger.Error.Printf("failed to save data point: %s", err)
						}
						l.Offset += int64(len(line)) + 1 // Adding 1 for \n
						l.Save()
					}

					line = line[:0]
					// Update the start index to be the next value.
					startIndex = i + 1

					continue
				}
			}

			// Add all data from the buffer
			line = append(line, buff[startIndex:n]...)
		}

		// Sleeping until the next slurp.
		time.Sleep(time.Duration(sleepTimeSeconds) * time.Second)
	}
}

// savePoint builds a new datapoint from line and sends it to the dataPointChan.
func savePoint(line []byte, dataPointChan chan<- device.DataPoint) error {
	d, err := device.ParseDataPoint(line)
	if err != nil {
		return fmt.Errorf("failed to build datapoint for saving: %s", err)
	}

	dataPointChan <- d

	return nil
}

// StartSlurp starts slurping the file if possible and sending data to the
// dataPointChan specified. sleepTimeSeconds is the amount of time the slurper
// sleeps before looking for new data in the file. The minimum value is 1.
func (l *LogFile) StartSlurp(dataPointChan chan<- device.DataPoint, sleepTimeSeconds int, maxShutdownWait float64) {
	// Don't start a new slurp if it's already running.
	if l.slurpRunning {
		logger.Verbose.Printf("slurper already started for %s", l.LogFilePath)
		return
	}
	go l.slurp(dataPointChan, sleepTimeSeconds)
	logger.Verbose.Printf("starting slurper for %s at offset %d", l.LogFilePath, l.Offset)
}

// StopSlurp stops slupring and blocks until stopped.
func (l *LogFile) StopSlurp() {
	if !l.SlurpRunning() {
		return
	}
	close(l.slurpCancelChan)
	start := time.Now()
	for l.SlurpRunning() {
		if time.Since(start).Seconds() > l.SlurperShutdownMaxWaitSeconds {
			logger.Verbose.Printf("forcing slurp of %s to stop due to exceeding limit", l.LogFilePath)
			return
		}
		logger.Debug.Printf("slurpRunning is: %v", l.SlurpRunning())
		time.Sleep(time.Duration(1) * time.Second)

	}
	logger.Verbose.Printf("slurper for %s has stopped: ", l.LogFilePath)
	return
}
