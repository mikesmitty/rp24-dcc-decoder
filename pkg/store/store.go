package store

import (
	"runtime"
	"time"

	"tinygo.org/x/tinyfs"
)

// CVFlags uses a bitmap for efficient storage and checking
type CVFlags uint8

const (
	Volatile   CVFlags = 0         // No flags
	Dirty      CVFlags = 1 << iota // CV has been written to
	ReadOnly                       // CV is read-only
	Persistent                     // CV should be saved to non-volatile memory
)

// Data holds the value, default, and flags for a single CV
type Data struct {
	Value   uint8
	Default uint8
	Flags   CVFlags
}

type Store struct {
	// Store is our map of CV number to CVData
	async bool
	// uint16 because 64kV ought to be enough for anybody
	data      map[uint16]Data
	fs        tinyfs.Filesystem
	index     uint16
	indexFile string
	ticker    *time.Ticker
}

// NewStore sets up the CV store and mounts the filesystem
func NewStore() *Store {
	c := &Store{
		data: make(map[uint16]Data, 256),
	}

	err := c.initFlash()
	if err != nil {
		println("could not initialize flash: " + err.Error())
	}

	return c
}

// Run periodically processes CV changes and persists them to flash
func (s *Store) Run() {
	s.ticker = time.NewTicker(500 * time.Millisecond)
	s.async = true
	for {
		select {
		case <-s.ticker.C:
			s.ProcessChanges()
		default:
			runtime.Gosched()
		}
	}
}

// SetDefault sets the value, default value and flags for a CV
func (s *Store) SetDefault(cvNumber uint16, defaultValue uint8, flags CVFlags) {
	s.data[cvNumber] = Data{
		Value:   defaultValue,
		Default: defaultValue,
		Flags:   flags,
	}
}

// SetReadOnly sets a CV to read-only
func (s *Store) SetReadOnly(cvNumber uint16) {
	data, ok := s.data[cvNumber]
	if !ok {
		return // CV not found or unused
	}
	data.Flags |= ReadOnly
	s.data[cvNumber] = data
}

// CV retrieves the value of a CV
func (s *Store) CV(cvNumber uint16) (uint8, bool) {
	data, ok := s.data[cvNumber]
	if !ok {
		return 0, false // CV not found or unused
	}
	return data.Value, true
}

// IndexedCV retrieves the value of a CV using the provided index
// Does not check whether or not the CV is valid, only if the read succeeded
func (s *Store) IndexedCV(index, cvNumber uint16) (uint8, bool) {
	if index == s.index {
		return s.CV(cvNumber)
	}
	return s.ReadCVFromFlash(index, cvNumber)
}

// Set sets the value of a CV, marking it as dirty if it's not read-only
func (s *Store) Set(cvNumber uint16, value uint8) bool {
	data, ok := s.data[cvNumber]
	if (data.Flags & ReadOnly) != 0 {
		return false // CV is read-only
	}
	if !ok || data.Value != value { // Only update if the value is different or key is missing
		data.Value = value
		data.Flags |= Dirty     // Mark as dirty
		s.data[cvNumber] = data // Store back into the map
	}

	// If we're not running async, persist the change immediately
	if !s.async && (data.Flags&Dirty) != 0 && (data.Flags&Persistent) != 0 {
		return s.persist(s.index, cvNumber, value)
	}
	return true
}

// Reset resets a single CV to its default value
func (s *Store) Reset(cvNumber uint16) bool {
	data, ok := s.data[cvNumber]
	if !ok {
		return false // CV not found or unused
	}
	if data.Value != data.Default { // Only update and mark dirty if needed
		data.Value = data.Default
		data.Flags |= Dirty
		s.data[cvNumber] = data
		// If we're not running async, persist the change immediately
		if !s.async {
			return s.persist(s.index, cvNumber, data.Value)
		}
	}
	return true
}

// ResetAllCVs resets all used CVs to their default values
func (s *Store) ResetAll() {
	for cvNumber := range s.data {
		s.Reset(cvNumber)
	}
}

// Clear empties the CV store in preparation for loading a new index
func (s *Store) Clear() {
	clear(s.data)
}

// Persist sets a CV value and immediately writes it to onboard flash
func (s *Store) Persist(cvNumber uint16, value uint8) bool {
	// Since this is for the current index, also set the current value
	s.Set(cvNumber, value)
	return s.persist(s.index, cvNumber, value)
}

// IndexedPersist writes a CV value to onboard flash using the provided index
func (s *Store) IndexedPersist(index, cvNumber uint16, value uint8) bool {
	return s.persist(index, cvNumber, value)
}
