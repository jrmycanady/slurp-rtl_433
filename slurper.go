package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jrmycanady/slurp-rtl_433/logger"
)

type FileRequest struct {
	FilePath              string
	Offset                int64
	MonitorTimeoutSeconds int64
	BufferSize            int
}

const (
	// The byte value for a linefeed.
	lf byte = 10

	// The byte value for a carraige return.
	cr byte = 13
)

// Slurper slurps files that are provided on the file channel and provides the results
// to the processor channel.
func slurper(fileReq FileRequest) {
	logger.Verbose.Printf("started slurper on %s", fileReq.FilePath)
	var buff = make([]byte, 10)
	var n int
	var sleepSeconds = 30
	var line = make([]byte, 0, 200)
	var offset int64

	// Opening the file for processing.
	f, err := os.Open(fileReq.FilePath)
	if err != nil {
		logger.Error.Printf("failed to open file %s", fileReq.FilePath)
		logger.Debug.Printf("failed to open file %s: %s", fileReq.FilePath, err)
		return
	}
	defer f.Close()

	logger.Verbose.Printf("opened %s for slurping", fileReq.FilePath)

	// Processing the file until somethign tells it to stop.
	for {
		// Seeking to the last location not recorded.
		_, err = f.Seek(offset, 0)
		if err != nil {
			logger.Error.Printf("failed to seek to file %s", fileReq.FilePath)
			logger.Debug.Printf("failed to seek to file %s: %s", fileReq.FilePath, err)
			return
		}
		logger.Debug.Printf("seeking complete on %s", fileReq.FilePath)

		// Reading until we reach the end of the file.
		for err != io.EOF {
			// Reading up to the buffer length.
			n, err = f.Read(buff)

			// Check each character in the buffer for line feed \n or carraige return \r.
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
						saveLine(line)
						// fmt.Printf("FOUND LINE: [%s]\n", string(line))
						offset += int64(len(line)) + 2 // Adding 2 for \r\n
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
					// save it. Then add anythign left to he new line.
					line = append(line, buff[startIndex:i]...)
					// Saving the line if not empty.
					if len(line) > 0 {
						saveLine(line)
						// fmt.Printf("FOUND LINE: [%s]\n", string(line))
						offset += int64(len(line)) + 1 // Adding 1 for \n
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
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}
}

func saveLine(line []byte) error {
	aw := AmbientWeather{}
	if err := json.Unmarshal(line, &aw); err != nil {
		panic(err)
	}
	aw.Parse()
	fmt.Println(aw)

	return nil
}
