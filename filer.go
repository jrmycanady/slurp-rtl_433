package main

import (
	"encoding/json"
	"io/ioutil"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jrmycanady/slurp-rtl_433/logger"
)

type LogFile struct {
	LastKnownFileName   string
	Offset              int
	Inode               uint64
	Birthtime           int64
	MTimeSpec           int64
	Size                int64
	ID                  uuid.UUID
	Lock                *sync.Mutex
	LogFileMetaDataPath string
	found               bool
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
	lf := LogFile{}

	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, err
	}
	lf.Lock = &sync.Mutex{}
	return &lf, nil
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
	}

	return &f
}

// FindLogFileByInode searches all known LogFiles for the inode provided.
func (f *Filer) FindLogFileByInode(inode uint64) *LogFile {

	for i := range f.Files {
		if f.Files[i].Inode == inode {
			return f.Files[i]
		}
	}
	return nil
}

// FindLogFiles searches the directory for log files creates new LogFile entries
// or updates current one to the found status.
func (f *Filer) FindLogFiles() error {
	logger.Verbose.Println("started find for log files")

	if !f.configured {
		panic("filer is not configured")
	}

	// Opening directory to get list of all possible files.
	files, err := ioutil.ReadDir(f.Config.dataFileDir)
	if err != nil {
		return err
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
			continue
		}

		// Building new file and adding to the list.
		uuid, err := uuid.NewRandom()
		if err != nil {
			panic(err)
		}
		newFile := LogFile{
			LastKnownFileName: files[i].Name(),
			Offset:            0,
			Birthtime:         stat.Birthtimespec.Sec,
			MTimeSpec:         stat.Mtimespec.Sec,
			Size:              stat.Size,
			ID:                uuid,
			Lock:              &sync.Mutex{},
			found:             true,
		}

		logger.Info.Printf("found new file %s with inode %d", files[i].Name(), stat.Ino)
		newFile.Save()
		f.Files[newFile.ID.String()] = &newFile
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
		case <-cancel:
			logger.Info.Println("cancel received, stopping")
			done <- struct{}{}
			return
		}
	}

}
