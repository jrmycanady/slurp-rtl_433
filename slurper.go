package main

import (
	"io"
	"os"

	"github.com/jrmycanady/slurp-rtl_433/logger"
)

type FileRequest struct {
	FilePath              string
	Offset                int64
	MonitorTimeoutSeconds int64
	BufferSize            int
}

const (
// var lf = []byte("\n")
// var cr = []byte("\r")
)

// Slurper slurps files that are provided on the file channel and provides the results
// to the processor channel.
func slurper(fileReqs <-chan FileRequest) {

	// var n int

	for fileReq := range fileReqs {
		var stop bool
		var buff = make([]byte, fileReq.BufferSize)

		// Opening the file for processing.
		f, err := os.Open(fileReq.FilePath)
		if err != nil {
			logger.Error.Printf("failed to open file %s", fileReq.FilePath)
			logger.Debug.Printf("failed to open file %s: %s", fileReq.FilePath, err)
			continue
		}
		defer f.Close()

		// Processing the file until somethign tells it to stop.
		for !stop {

			// Seeking to the last location not recorded.
			_, err = f.Seek(fileReq.Offset, 0)
			if err != nil {
				logger.Error.Printf("failed to seek to file %s", fileReq.FilePath)
				logger.Debug.Printf("failed to seek to file %s: %s", fileReq.FilePath, err)
				continue
			}

			// LINELOOP:
			// Reading until we reach the end of the file.
			for err != io.EOF {
				_, err = f.Read(buff)
			}
		}
	}
}
