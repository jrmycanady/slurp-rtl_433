package file

import (
	"testing"
)

func TestValidateLogFileName(t *testing.T) {
	configName := "rtl_433_data.log"

	goodName := "rtl_433_data.log"
	if !validateLogFileName(configName, goodName) {
		t.Fatalf("goodName failed")
	}

	goodNameRotated := "rtl_433_data.log.1"
	if !validateLogFileName(configName, goodNameRotated) {
		t.Fatalf("goodNameRotated failed")
	}

	badName := "blarg"
	if validateLogFileName(configName, badName) {
		t.Fatalf("badName succeeded")
	}

	// configNameWithPath := "/path/to/file/rtl_433_data.log"

}
