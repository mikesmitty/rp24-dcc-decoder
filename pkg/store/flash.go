//go:build rp

package store

import (
	"io"
	"machine"
	"os"
	"strconv"

	"tinygo.org/x/tinyfs/littlefs"
)

func (s *Store) initFlash() error {
	// FIXME: Make sure to format the filesystem so there is room for the code to grow

	// Set up the filesystem
	fs := littlefs.New(machine.Flash)
	fs.Configure(&littlefs.Config{
		CacheSize:     512,
		LookaheadSize: 512,
		BlockCycles:   100,
	})
	s.fs = fs

	err := s.fs.Mount()
	if err != nil {
		// If the filesystem is un-recoverable, format it
		if !s.recoverableMountError(err) {
			println("filesystem unrecoverable, reformatting")
			if err := s.fs.Format(); err != nil {
				println("could not format filesystem: " + err.Error())
				return err
			}
			if err := s.fs.Mount(); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	info, err := s.fs.Stat("cvstore")
	if err != nil && err.Error() != "littlefs: No directory entry" {
		return err
	}
	if info == nil {
		// The FileMode is ignored in TinyFS
		if err := s.fs.Mkdir("cvstore", 0x755); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) recoverableMountError(err error) bool {
	switch err.Error() {
	case "littlefs: Corrupted":
		return false
	case "littlefs: Invalid parameter":
		return false
	}
	return true
}

// persist immediately writes a CV value to onboard flash
func (s *Store) persist(index, cvNumber uint16, value uint8) bool {
	if cvNumber < 1 || cvNumber > 256 {
		println("CV number out of range")
		return false
	}

	indexFile := indexFileName(index)

	f, err := s.fs.OpenFile(indexFile, os.O_WRONLY)
	if err != nil {
		println("could not open file: " + err.Error())
		return false
	}
	_, err = f.Seek(int64(cvNumber-1), io.SeekStart)
	if err != nil {
		println("could not seek file: " + err.Error())
		return false
	}
	_, err = f.Write([]byte{value})
	if err != nil {
		println("could not write to file: " + err.Error())
		return false
	}
	err = f.(*littlefs.File).Sync()
	if err != nil {
		println("could not sync writes to '", indexFile, "': "+err.Error())
		return false
	}
	err = f.Close()
	if err != nil {
		println("could not close file: " + err.Error())
		return false
	}

	return true
}

// ReadCVFromFlash loads a CV value from flash
// Does not check whether or not the CV is valid, only if the read succeeded
func (s *Store) ReadCVFromFlash(index, startCV uint16) (uint8, bool) {
	v, ok := s.ReadCVsFromFlash(index, startCV, 1)
	if len(v) == 0 {
		return 0, false
	}
	return v[0], ok
}

// ReadCVsFromFlash loads a list of CVs from flash
// Does not check whether or not the CV is valid, only if the read succeeded
func (s *Store) ReadCVsFromFlash(index, startCV, count uint16) ([]uint8, bool) {
	// Open the index file
	f, err := s.fs.OpenFile(indexFileName(index), os.O_RDONLY)
	if err != nil {
		println("could not open file: " + err.Error())
		return nil, false
	}

	// Read the CVs from the index
	buf := make([]byte, count)
	_, err = f.Seek(int64(startCV-1), io.SeekStart)
	if err != nil {
		println("could not seek file: " + err.Error())
		return nil, false
	}
	_, err = f.Read(buf)
	if err != nil {
		println("could not read from file: " + err.Error())
		return nil, false
	}

	return buf, true
}

// LoadIndex loads a persistent CV index from flash
// Bool return value indicates if the index file was found
func (s *Store) LoadIndex(newIndex uint16) (bool, error) {
	newIndexFile := indexFileName(newIndex)

	// Check if the index file exists
	ok := false
	if _, err := s.fs.Stat(newIndexFile); err == nil {
		ok = true
	}

	// Open the index file
	f, err := s.fs.OpenFile(newIndexFile, os.O_RDWR|os.O_CREATE)
	if err != nil {
		return false, err
	}

	// If the file already exists, load the CVs we have been informed of by SetDefault()
	if ok {
		buf := []byte{0}
		for cvNumber, data := range s.data {
			if cvNumber < 1 {
				continue
			}
			if (data.Flags & Persistent) != 0 {
				// Read the CV from flash
				_, err = f.Seek(int64(cvNumber-1), io.SeekStart)
				if err != nil {
					println("could not seek file: " + err.Error())
					return false, nil
				}
				_, err = f.Read(buf)
				if err != nil {
					println("could not read from file: " + err.Error())
					return false, nil
				}
			}
		}
	} else {
		// If the file is new, pad it out to 256 bytes
		_, err := f.Write(make([]byte, 256))
		if err != nil {
			return false, err
		}
		err = f.(*littlefs.File).Sync()
		if err != nil {
			return false, err
		}
	}

	// After successfully loading the index, set the index and index file
	s.index = newIndex
	s.indexFile = indexFileName(newIndex)

	return ok, nil
}

// indexFileName returns the filename for the given index file
func indexFileName(index uint16) string {
	return "cvstore/index" + strconv.FormatUint(uint64(index), 10) + ".bin"
}

// ProcessChanges iterates through the current CV index and persists any dirty CVs
func (s *Store) ProcessChanges() bool {
	f, err := s.fs.OpenFile(s.indexFile, os.O_WRONLY)
	if err != nil {
		println("could not open file: " + err.Error())
		return false
	}

	// Write all the changed CVs to flash
	for cvNumber, data := range s.data {
		if (data.Flags & Dirty) != 0 {
			if (data.Flags & Persistent) != 0 {
				_, err = f.Seek(int64(cvNumber-1), io.SeekStart)
				if err != nil {
					println("could not seek file: " + err.Error())
					return false
				}
				_, err = f.Write([]byte{data.Value})
				if err != nil {
					println("could not write to file: " + err.Error())
					return false
				}
			}
		}
	}

	// Sync and close the file
	err = f.(*littlefs.File).Sync()
	if err != nil {
		println("could not sync writes to '", s.indexFile, "': "+err.Error())
		return false
	}
	err = f.Close()
	if err != nil {
		println("could not close file: " + err.Error())
		return false
	}

	// Now that the file is saved clear the dirty flags
	for cvNumber, data := range s.data {
		if (data.Flags & Dirty) != 0 {
			// Clear the dirty flag
			data.Flags &^= Dirty    // Use bitwise AND NOT to clear the flag
			s.data[cvNumber] = data // Store back into the map
		}
	}

	return true
}
