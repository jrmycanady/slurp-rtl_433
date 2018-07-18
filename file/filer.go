package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jrmycanady/slurp-rtl_433/config"
	"github.com/jrmycanady/slurp-rtl_433/device"
	"github.com/jrmycanady/slurp-rtl_433/logger"
)

const (
	dot byte = 46
)

var (
	logFileRE = regexp.MustCompile(`^(?P<filename>.*)(\.\d*)?\.log$`)
)

// A Filer manages finding/identifying log files as well as starting/stopping
// the slurping of the files.
//
// Filers must be configured before they can be started. This can be done during
// creation using NewFiler(..)
type Filer struct {
	// Files contains all the files currently known to the filer.
	Files map[string]*LogFile

	// cfg is the configuration to filer will use.
	cfg config.Config

	// Configured is true if the filer has been configured. A filer will not
	// start if it has not been configured.
	configured bool

	// Running is true if the filer is currently running. A filer will not
	// restart if already running.
	running bool

	// CancelChan provides access to the cancel channel. By closing the channel
	// the Filer will begin the cancel process and stop. Generally Stop() should
	// be used but the channel is available if needed.
	CancelChan chan struct{}

	// donechan receives a message when the filer is done. This is used to
	// determine when the filer is done.
	doneChan chan struct{}

	// dropOff
	dropOffChan chan<- device.DataPoint
}

// Running returns the current running status of the Filer.
func (f *Filer) Running() bool {
	return f.running
}

// Configured returns the current configuration status of the Filer.
func (f *Filer) Configured() bool {
	return f.configured
}

// Config returns the current configuration fo the Filer. An empty
// configuration is returned if the Filer is not configured. Use Configured()
// to determine if the filer has been configured.
func (f *Filer) Config() config.Config {
	return f.cfg
}

// NewFiler creates a new Filer and builds all the required internal channels.
// dropOffChan should be a channel that is monitored for DataPoints and then
// processed as needed.
func NewFiler(cfg config.Config, dropOffChan chan<- device.DataPoint) *Filer {
	return &Filer{
		Files:       make(map[string]*LogFile),
		cfg:         cfg,
		configured:  true,
		CancelChan:  make(chan struct{}),
		doneChan:    make(chan struct{}, 2),
		dropOffChan: dropOffChan,
	}
}

// findLogFileByInode searches all known LogFiles for the inode provided.
func (f *Filer) findLogFileByInode(inode uint64) *LogFile {

	logger.Debug.Println("searching for file by inode")
	for i := range f.Files {
		logger.Debug.Printf("%d =?= %d", f.Files[i].Inode, inode)
		if f.Files[i].Inode == inode {
			logger.Debug.Printf("found known file with inode %d", inode)
			return f.Files[i]
		}
	}
	logger.Debug.Printf("did not find known file with inode %d", inode)
	return nil
}

// Start starts the Filer process to look for and slurp log files. If the
// Filer has not been configured it will panic as starting a filer that is
// not configured is a design error.
func (f *Filer) Start() error {
	var err error

	logger.Info.Println("starting filer")

	// Doing nothing if not configured.
	if !f.configured {
		panic(fmt.Errorf("cannot start filer that is not configured"))

	}

	// Loading known log files from the meta data.
	if err = f.loadLogMetaData(); err != nil {
		logger.Error.Printf("failed to load let meta data files: %s", err)
		f.shutdown()
		return fmt.Errorf("failed to start filer: %s", err)
	}

	logger.Info.Printf("filer found %d log meta data files", len(f.Files))

	// Starting process loop.
	go f.run()
	f.running = true

	return nil
}

// Stop issues a cancel to all slurpers and will block until everything is
// not running.
func (f *Filer) Stop() {
	// Closing the cancel channel to send the notice.
	close(f.CancelChan)

	// Waiting for the status to update or the max wait time to be exceeded.
	start := time.Now()
	for {
		// No longer running to exit.
		if f.running == false {
			return
		}

		time.Sleep(time.Duration(1) * time.Second)

		if time.Since(start).Seconds() > float64(f.cfg.FilerShutdownMaxWaitSeconds) {
			logger.Error.Printf("exceeded FilerShutdownMaxWaitTimeSeconds of %d, forcing shutdown now", f.cfg.FilerShutdownMaxWaitSeconds)
			return
		}
	}

}

// run is the primary working loop of the filer. It should be executed as a
// go routine in most cases.
func (f *Filer) run() {
	var err error

	// Generating the ticker to for checks for new files.
	findTimer := time.NewTicker(time.Duration(f.cfg.LogFileCheckTimeSeconds) * time.Second)

	for {
		select {
		case <-findTimer.C:
			if err = f.findAndSlurpLogFiles(); err != nil {
				logger.Error.Printf("failed to find any log files: %s", err)
			} else {
				logger.Info.Println("log file search complete")
			}
		case <-f.CancelChan:
			logger.Info.Println("cancel received, stopping all file slurpers")

			for i := range f.Files {
				logger.Verbose.Printf("stopping slurper for %s", f.Files[i].LogFilePath)
				f.Files[i].StopSlurp()
				logger.Debug.Printf("stopping of slurper for %s compelte", f.Files[i].LogFilePath)
			}
			f.shutdown()
			logger.Info.Println("filer has stopped")
			return
		}
	}
}

// shutdown sets the filer to shutdown status. It does not wait for anything to
// stop. Generally Stop() should be used.
func (f *Filer) shutdown() {
	f.doneChan <- struct{}{}
	f.running = false
	return
}

func (f *Filer) loadLogMetaData() error {
	if !f.configured {
		panic(fmt.Errorf("cannot start filer that is not configured"))
	}

	// Opening the metadata directory and loading all metadata files.
	files, err := ioutil.ReadDir(f.cfg.FileMetaDataPath)
	if err != nil {
		return err
	}

	for i := range files {

		// Ignoring all directories and any files with a length too short.
		if files[i].IsDir() || len(files[i].Name()) != 41 {
			logger.Info.Printf("metadata load ignoring directory or file with too short of name %s", files[i].Name())
			continue
		}

		// Parsing the uuid so we can build the file.
		id, err := uuid.Parse(files[i].Name()[:36])
		if err != nil {
			logger.Info.Printf("meta data file name did not parse into uuid correctly: %s", files[i].Name())
			continue
		}

		// Read all the data in the file.
		file, err := ioutil.ReadFile(f.cfg.FileMetaDataPath + "/" + id.String() + ".meta")
		if err != nil {
			continue
		}

		// Load the metadate file.
		lf, err := NewLogFile(file)
		if err != nil {
			logger.Info.Printf("failed to load meta data from %s", files[i].Name())
			continue
		}

		f.Files[id.String()] = lf

	}

	return nil
}

// findAndSlurpLogFiles searches the log file directory for log files and
// begins the slurp process on any that need to be processed. It's assumed that
// the Slurp method of the LogFile is smart enough to handle multiple calls
// and will not double slurp or slurp an old file.
func (f *Filer) findAndSlurpLogFiles() error {
	var err error
	logger.Verbose.Printf("starting find for new log files for slurping")

	// Opening directory to get list of all possible files for slurping.
	files := make([]os.FileInfo, 0, 0)
	if f.cfg.DataFileDir != "" {
		files, err = ioutil.ReadDir(f.cfg.DataFileDir)
		if err != nil {
			return fmt.Errorf("failed to read directory at %s: %s", f.cfg.DataFileDir, err)
		}
	} else {
		// No directory was in the config so it must be a single file. If no
		// file is found it's not an error as it could show up later.
		file, err := os.Stat(f.cfg.DataFileDir)
		if err != nil {
			logger.Verbose.Printf("did not find log file named %s", f.cfg.DataFileDir)
			return nil
		}
		files = append(files, file)
	}

	// Checking each file to see if it's a log file.
	for i := range files {
		logger.Verbose.Printf("checking (file | dir) %s", files[i].Name())

		// Validating the name matches the expected file name.
		if !validateLogFileName(f.cfg.DataFileName, files[i].Name()) {
			logger.Verbose.Printf("%s name does not match expected format %s", files[i].Name(), f.cfg.DataFileName)
			continue
		}

		if files[i].IsDir() {
			// Ignoring any directories.
			logger.Verbose.Printf("%s is a directory, ignoring", files[i].Name())
			continue
		}

		stat, ok := files[i].Sys().(*syscall.Stat_t)
		if !ok {
			// Ignoring any files we can't stat.
			logger.Verbose.Printf("failed to stat %s, ignoring", files[i].Name())
			continue
		}

		// Finding if we already have the file and processing accordinly.
		foundFile := f.findLogFileByInode(stat.Ino)
		if foundFile != nil {
			logger.Verbose.Printf("file %s already known to filer, updating to found", files[i].Name())
			foundFile.SetFound(true)
			foundFile.StartSlurp(f.dropOffChan, f.cfg.SlurpSleepTimeSeconds)
			continue
		}

		// Building new file and adding to the list.
		newFile, err := NewLogFile([]byte{})
		if err != nil {
			logger.Error.Printf("failed to load file %s: %s", files[i].Name(), err)
			continue
		}

		newFile.LogFilePath = fmt.Sprintf("%s/%s", f.cfg.DataFileDir, files[i].Name())
		newFile.Offset = 0
		newFile.Inode = stat.Ino
		newFile.found = true
		newFile.MetaDataFilePath = fmt.Sprintf("%s/%s.meta", f.cfg.FileMetaDataPath, newFile.MetaDataID)

		logger.Info.Printf("found new file %s with inode %d", files[i].Name(), stat.Ino)
		if err = newFile.Save(); err != nil {
			logger.Info.Printf("failed to save file meta data to %s, not slurping", newFile.MetaDataFilePath)
			continue
		}
		f.Files[newFile.MetaDataID.String()] = newFile
		newFile.StartSlurp(f.dropOffChan, f.cfg.SlurpSleepTimeSeconds)
	}
	return nil

}

// validateLogFileName validates the found log file name matches the expected
// name taking into account logrotate number indicators. It returns true if
// the name is valid otherwise it will return false. It will also return false
// if the expected name does not end with .log.
func validateLogFileName(expected string, found string) bool {
	expName := expected[:len(expected)-4]
	expExt := expected[len(expected)-4:]

	foundName := found[:len(found)-4]
	foundExt := found[len(found)-4:]

	// Validating both files have proper extension.
	if expExt != ".log" || foundExt != ".log" {
		fmt.Println("not .log")
		return false
	}

	// Validate it's long enough.
	if len(expName) > len(foundName) {
		return false
	}

	// Validate name matches
	if expName != foundName[:len(expName)] {
		return false
	}

	if len(foundName) > len(expName) {
		tick := foundName[len(expName):]
		if tick[0] != dot {
			return false
		}

		_, err := strconv.Atoi(tick[1:])
		if err != nil {
			return false
		}

	}

	return true
}
