// Package logger is a logging package that allow some granularity in logging by
// building separate loggers for each category. The output location of
// the logs can be specified on creation.
//
// Logging Levels
//
// The logger offers 4 log levels. In general Error and Info should be used
// if at all possible.
// Error:
//  * Fails to perform an action that should be successful. i.e. Database failures.
// Info:
//  * General program flow used by administrators
// Verbose:
//  * Detailed flow that may be used to track production issues.
// Debug:
//  * Very low level state items that should never hit log files in production.
//  * Sensitive information could be in the log file..
//  * Generally these could be removed without any end user ever carring
//
// Usage
//
// The logger configures on import to utilize os.Stdout. By default all
// loggers are enabled. The output location can be modified directly on
// the variable as they are normal log instances. To disable a log
// simply use SetOutput.
package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	// Error is a logger that logs on error level.
	Error = log.New(os.Stdout, "  [ERROR] ", log.Ldate|log.Ltime|log.LUTC)

	// Info is the logger that logs on the info level.
	Info = log.New(os.Stdout, "   [INFO] ", log.Ldate|log.Ltime|log.LUTC)

	// Verbose is a logger that logs at the verbose level.
	Verbose = log.New(os.Stdout, "[VERBOSE] ", log.Ldate|log.Ltime|log.LUTC)

	// Debug is a logger that logs at the debug level.
	Debug = log.New(os.Stdout, "  [DEBUG] ", log.Ldate|log.Ltime|log.LUTC)
)

// Update takes an io.Writer and depending on if enable is true will set the logger
// to the io.Writer or in the case of false set it to Discard.
// This can be used to setup the logging level at a global scale.
func Update(handler io.Writer, enableError bool, enableInfo bool, enableVerbose, bool, enableDebug bool) {
	if enableError {
		Error.SetOutput(handler)
	} else {
		Error.SetOutput(ioutil.Discard)
	}

	if enableInfo {
		Info.SetOutput(handler)
	} else {
		Info.SetOutput(ioutil.Discard)
	}

	if enableVerbose {
		Verbose.SetOutput(handler)
	} else {
		Verbose.SetOutput(ioutil.Discard)
	}

	if enableDebug {
		Debug.SetOutput(handler)
	} else {
		Debug.SetOutput(ioutil.Discard)
	}
}

// UpdateWithLevelList updates the handler of each level to handler based on the levels provided.
func UpdateWithLevelList(handler io.Writer, errorLevels []string) {

	// Configure all but error to discard by default.
	Info.SetOutput(ioutil.Discard)
	Verbose.SetOutput(ioutil.Discard)
	Debug.SetOutput(ioutil.Discard)

	for _, s := range errorLevels {
		switch s {
		case "error":
			Error.SetOutput(handler)
		case "info":
			Info.SetOutput(handler)
		case "verbose":
			Verbose.SetOutput(handler)
		case "debug":
			Debug.SetOutput(handler)
		}
	}
}

// ConfigureWithFile configures the output to use the file at the path provided.
func ConfigureWithFile(filePath string, logLevels []string) (*os.File, error) {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file for writing at %v: %v", filePath, err)
	}

	UpdateWithLevelList(f, logLevels)
	return f, err

}

// EventLevel is the logging level the even should have.
type EventLevel int

const (
	// EventLevelInfo represents an event that should be logged at the info level.
	EventLevelInfo EventLevel = 0

	// EventLevelError represents an event that should be logged at the error level.
	EventLevelError EventLevel = 1

	// EventLevelDebug represents an event that should be logged at the debug level.
	EventLevelDebug EventLevel = 2

	// EventLevelVerbose represents an event that should be logged at the verbose level.
	EventLevelVerbose EventLevel = 3
)
