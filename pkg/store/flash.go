//go:build rp

package store

import (
	"io"
	"machine"
	"os"

	"tinygo.org/x/tinyfs/littlefs"
)

func (c *Store) initFlash() error {
	// Set up the filesystem
	fs := littlefs.New(machine.Flash)
	fs.Configure(&littlefs.Config{
		CacheSize:     512,
		LookaheadSize: 512,
		BlockCycles:   100,
	})
	c.fs = fs

	err := c.fs.Mount()
	if err != nil {
		// If the filesystem is corrupted (or blank), format it
		if err.Error() == "littlefs: Corrupted" {
			if err := c.fs.Format(); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if err := c.fs.Mount(); err != nil {
		return err
	}

	return nil
}

// persist immediately writes a CV value to onboard flash
func (c *Store) persist(index, cvNumber uint16, value uint8) bool {
	if cvNumber < 1 || cvNumber > 256 {
		println("CV number out of range")
		return false
	}

	indexFile := indexFileName(index)

	f, err := c.fs.OpenFile(indexFile, os.O_WRONLY)
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
func (c *Store) ReadCVFromFlash(index, startCV uint16) (uint8, bool) {
	v, ok := c.ReadCVsFromFlash(index, startCV, 1)
	if len(v) == 0 {
		return 0, false
	}
	return v[0], ok
}

// ReadCVsFromFlash loads a list of CVs from flash
// Does not check whether or not the CV is valid, only if the read succeeded
func (c *Store) ReadCVsFromFlash(index, startCV, count uint16) ([]uint8, bool) {
	// Open the index file
	f, err := c.fs.OpenFile(indexFileName(index), os.O_RDONLY)
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
func (c *Store) LoadIndex(newIndex uint16, cvs []byte) (bool, error) {
	newIndexFile := indexFileName(newIndex)

	// Check if the index file exists
	ok := false
	if _, err := c.fs.Stat(newIndexFile); err == nil {
		ok = true
	}

	// Open the index file
	f, err := c.fs.OpenFile(newIndexFile, os.O_RDWR|os.O_CREATE)
	if err != nil {
		return false, err
	}

	// Clear the existing CV data
	clear(c.data)

	if ok {
		// If the file exists, read the index
		buf := make([]byte, 256)
		_, err := f.Read(buf)
		if err != nil {
			return true, err
		}
		// Load the CVs from the index as indicated by the provided bitmap
		for i := range cvs {
			for j := 0; j < 8; j++ {
				if cvs[i]&(1<<j) != 0 {
					c.data[uint16(i*8+j)] = Data{
						Value:   buf[i*8+j],
						Default: buf[i*8+j],
						Flags:   Persistent,
					}
				}
			}
		}
	} else {
		// If the file is new, pad it out to 256 bytes
		for i := 0; i < 256; i++ {
			f.Write([]byte{0})
		}
		err := f.(*littlefs.File).Sync()
		if err != nil {
			return false, err
		}

		// Initialize the CVs as indicated by the provided bitmap
		for i := range cvs {
			for j := 0; j < 8; j++ {
				if cvs[i]&(1<<j) != 0 {
					c.data[uint16(i*8+j)] = Data{
						Flags: Persistent,
					}
				}
			}
		}
	}

	// After successfully loading the index, set the index and index file
	c.index = newIndex
	c.indexFile = indexFileName(newIndex)

	return ok, nil
}

// indexFileName returns the filename for the given index file
func indexFileName(index uint16) string {
	return "cvstore/index" + string(index+0x30) + ".bin"
}
