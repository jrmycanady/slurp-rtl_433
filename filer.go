package main

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
	"github.com/jrmycanady/slurp-rtl_433/logger"
)

// A LogFile represents a lot file that is slurped for data.
type LogFile struct {
	LastKnownFilePath   string
	Offset              int64
	Inode               uint64
	Birthtime           int64
	MTimeSpec           int64
	Size                int64
	ID                  uuid.UUID
	Lock                *sync.Mutex
	LogFileMetaDataPath string
	found               bool
	slurpRunning        bool
	slurpCancel         chan struct{}
	slurpError          error
}

// StartSlurp starts a slurp goroutine on the file if possible. If one
// is already running based on LogFile.slurRunning then a new one
// is not started.
func (l *LogFile) StartSlurp() {
	// Don't start a new slurp if it's already running.
	if l.slurpRunning {
		return
	}
	go l.slurp()
}

// StopSlurp tells the slurper to stop.
func (l *LogFile) StopSlurp() {
	close(l.slurpCancel)
}

func (l *LogFile) setSlurpRunning(running bool) {
	l.Lock.Lock()
	l.slurpRunning = running
	l.Lock.Unlock()
}

// GetSlurpStatus returns the current status of the slurper.
func (l *LogFile) GetSlurpStatus() bool {
	return l.slurpRunning
}

func (l *LogFile) slurp() {
	l.setSlurpRunning(true)
	defer l.setSlurpRunning(false)

	// Opening the file for processing.
	f, err := os.Open(l.LastKnownFilePath)
	if err != nil {
		logger.Error.Printf("failed to open file %s", l.LastKnownFilePath)
		logger.Debug.Printf("failed to open file %s: %s", l.LastKnownFilePath, err)
		return
	}
	defer f.Close()

	// Validate it's still the same file.
	stat, err := f.Stat()
	if err != nil {
		logger.Error.Printf("failed to stat file %s", l.LastKnownFilePath)
		return
	}
	sys := stat.Sys().(*syscall.Stat_t)
	if sys.Ino != l.Inode {
		logger.Error.Printf("the inode has changed so the file is new %s", l.LastKnownFilePath)
		return
	}

	logger.Info.Printf("opened and starting slruping of %s", l.LastKnownFilePath)

	// Processing the file until something tells it to stop.
	for {
		var n int
		var buff = make([]byte, 10)
		var line = make([]byte, 0, 200)

		// Seeking to the last location not recorded.
		_, err = f.Seek(l.Offset, 0)
		if err != nil {
			logger.Error.Printf("failed to seek to file %s", l.LastKnownFilePath)
			logger.Debug.Printf("failed to seek to file %s: %s", l.LastKnownFilePath, err)
			return
		}
		logger.Debug.Printf("seeking complete on %s", l.LastKnownFilePath)

		// Reading until we reach the end of the file.
		for err != io.EOF {
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
						saveLine(line)

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
						saveLine(line)
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
		// Check to see if we should stop.
		select {
		case <-l.slurpCancel:
			return
		default:
		}

		time.Sleep(time.Duration(30) * time.Second)
	}

}

// NewLogFile creates a new LogFile with working mutexes and channels, and a UUID that may be replaced.
func NewLogFile() *LogFile {
	uuid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	return &LogFile{
		ID:          uuid,
		Lock:        &sync.Mutex{},
		slurpCancel: make(chan struct{}),
	}
}

// SetID sets the id of the log file.
func (l *LogFile) SetID(id uuid.UUID) {
	l.Lock.Lock()
	l.ID = id
	l.Lock.Unlock()
}

// Found returns if the LogFile was found on the last discovery.
func (l *LogFile) Found() bool {
	return l.found
}

// SetFound sets the found status of the LogFile.
func (l *LogFile) SetFound(f bool) {
	l.Lock.Lock()
	l.found = f
	l.Lock.Unlock()
}

// LoadLogFileFromJSON creates a LogFile from based on the data provided in json
// format. It also creates a new working mutex.
func LoadLogFileFromJSON(data []byte) (*LogFile, error) {
	lf := NewLogFile()

	if err := json.Unmarshal(data, lf); err != nil {
		return nil, err
	}
	// lf.Lock = &sync.Mutex{}
	return lf, nil
}

// Save saves the log file data.
func (l *LogFile) Save() error {
	l.Lock.Lock()
	defer l.Lock.Unlock()
	j, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(l.LogFileMetaDataPath, j, 0644)

}

// The Filer represents the process that manages all the files
// and slurpers.
type Filer struct {
	Files      map[string]*LogFile
	Config     Config
	configured bool
}

// NewFiler returns a newley configured filer.
func NewFiler(config Config) *Filer {
	f := Filer{
		Config:     config,
		configured: true,
		Files:      make(map[string]*LogFile),
	}

	return &f
}

// FindLogFileByInode searches all known LogFiles for the inode provided.
func (f *Filer) FindLogFileByInode(inode uint64) *LogFile {

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

// FindLogFiles searches the directory for log files creates new LogFile entries
// or updates current one to the found status.
func (f *Filer) FindLogFiles() error {
	var err error
	logger.Verbose.Println("started find for log files")

	if !f.configured {
		panic("filer is not configured")
	}

	// Opening directory to get list of all possible files.
	files := make([]os.FileInfo, 0, 0)
	if f.Config.dataFileDir != "" {
		files, err = ioutil.ReadDir(f.Config.dataFileDir)
		if err != nil {
			return fmt.Errorf("failed to read directory at %s: %s", f.Config.dataFileDir, err)
		}
	} else {
		// Returning nil if the file is not found as it's not a true failure case.
		file, err := os.Stat(f.Config.dataFileName)
		if err != nil {
			logger.Verbose.Printf("did not find log file named %s", f.Config.dataFileName)
			return nil
		}
		files = append(files, file)
	}

	// Checking each file to see if it's a log file.
	for i := range files {
		logger.Verbose.Printf("checking (file | dir) %s", files[i].Name())

		if files[i].IsDir() {
			logger.Verbose.Printf("%s is a directory, ignoring", files[i].Name())
			// Ignoring any directories
			continue
		}

		stat, ok := files[i].Sys().(*syscall.Stat_t)
		if !ok {
			logger.Verbose.Printf("failed to stat %s, ignoring", files[i].Name())
			// Ignoring any file that we can't stat.
			continue
		}

		// Finding if we already have the file and processing accordinly.
		foundFile := f.FindLogFileByInode(stat.Ino)
		if foundFile != nil {
			logger.Info.Printf("already know of file %s, updating to found", files[i].Name())
			foundFile.SetFound(true)
			foundFile.StartSlurp()
			continue
		}

		// TODO validate name fo the file!

		// Building new file and adding to the list.
		uuid, err := uuid.NewRandom()
		if err != nil {
			panic(err)
		}
		newFile := NewLogFile()
		// newFile := LogFile{
		// 	LastKnownFileName:   files[i].Name(),
		// 	Offset:              0,
		// 	Birthtime:           stat.Birthtimespec.Sec,
		// 	MTimeSpec:           stat.Mtimespec.Sec,
		// 	Size:                stat.Size,
		// 	Inode:               stat.Ino,
		// 	ID:                  uuid,
		// 	Lock:                &sync.Mutex{},
		// 	found:               true,
		// 	LogFileMetaDataPath: fmt.Sprintf("%s/%s.meta", f.Config.FileMetaDataPath, uuid.String()),
		// }
		newFile.LastKnownFilePath = fmt.Sprintf("%s/%s", f.Config.dataFileDir, files[i].Name())
		newFile.Offset = 0
		newFile.Birthtime = stat.Birthtimespec.Sec
		newFile.MTimeSpec = stat.Mtimespec.Sec
		newFile.Size = stat.Size
		newFile.Inode = stat.Ino
		newFile.ID = uuid
		newFile.Lock = &sync.Mutex{}
		newFile.found = true
		newFile.LogFileMetaDataPath = fmt.Sprintf("%s/%s.meta", f.Config.FileMetaDataPath, uuid.String())

		logger.Info.Printf("found new file %s with inode %d", files[i].Name(), stat.Ino)
		if err = newFile.Save(); err != nil {
			logger.Info.Printf("failed to save file meta data to %s, not slurping", newFile.LogFileMetaDataPath)
			continue
		}
		f.Files[newFile.ID.String()] = newFile
		newFile.StartSlurp()
	}
	return nil
}

// LoadKnownLogFiles loads the log files from the meta data files..
func (f *Filer) LoadKnownLogFiles() error {
	if !f.configured {
		panic("filer is not configured")
	}

	// Opening the metadata directory and loading all metadata files.
	files, err := ioutil.ReadDir(f.Config.FileMetaDataPath)
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
		file, err := ioutil.ReadFile(f.Config.FileMetaDataPath + "/" + id.String() + ".meta")
		if err != nil {
			continue
		}

		// Load the metadate file.
		lf, err := LoadLogFileFromJSON(file)
		if err != nil {
			logger.Info.Printf("failed to load meta data from %s", files[i].Name())
			continue
		}

		f.Files[id.String()] = lf

	}

	return nil
}

// Start starts the filer which looks for log files and manages the slurpers for each.
func (f *Filer) Start(cancel <-chan struct{}, done chan<- struct{}) {
	var err error
	if !f.configured {
		panic("filer not configured")
	}

	// Load any known files.
	if err = f.LoadKnownLogFiles(); err != nil {
		logger.Error.Printf("failed to load log files: %s", err)
		done <- struct{}{}
		return
	}
	logger.Info.Printf("found %d known log files", len(f.Files))

	findTimer := time.NewTicker(time.Duration(f.Config.LogFileCheckTimeSeconds) * time.Second)

	for {
		select {
		case <-findTimer.C:
			logger.Info.Println("starting search for new log files")
			if err = f.FindLogFiles(); err != nil {
				logger.Error.Printf("failed to find log files: %s", err)
			} else {
				logger.Info.Println("search for log files completed")
			}

		case <-cancel:
			logger.Info.Println("cancel received, stopping")

			done <- struct{}{}
			return
		}
	}

}
