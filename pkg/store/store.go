package store

import (
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
	// uint16 because 64kV ought to be enough for anybody
	data      map[uint16]Data
	fs        tinyfs.Filesystem
	index     uint16
	indexFile string
}

/* FIXME: Implement this?
- After Delay:  You could use a time.Timer to implement a delay, resetting the timer on each CV write.
- On Demand:  This would be triggered by your DCC command parsing logic.
- Combination:  Combine the ticker with a "last write time" check.  This is the most robust approach.

func StartCVProcessingTimer() {
	ticker := time.NewTicker(50 * time.Millisecond)
	go func() {
		for range ticker.C {
			ProcessCVChanges()
		}
	}()
}
*/

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

// SetDefault sets the value, default value and flags for a CV
func (c *Store) SetDefault(cvNumber uint16, defaultValue uint8, flags CVFlags) {
	c.data[cvNumber] = Data{
		Value:   defaultValue,
		Default: defaultValue,
		Flags:   flags,
	}
}

// SetReadOnly sets a CV to read-only
func (c *Store) SetReadOnly(cvNumber uint16) {
	data, ok := c.data[cvNumber]
	if !ok {
		return // CV not found or unused
	}
	data.Flags |= ReadOnly
	c.data[cvNumber] = data
}

// CV retrieves the value of a CV
func (c *Store) CV(cvNumber uint16) (uint8, bool) {
	data, ok := c.data[cvNumber]
	if !ok {
		return 0, false // CV not found or unused
	}
	return data.Value, true
}

// IndexedCV retrieves the value of a CV using the provided index
// Does not check whether or not the CV is valid, only if the read succeeded
func (c *Store) IndexedCV(index, cvNumber uint16) (uint8, bool) {
	if index == c.index {
		return c.CV(cvNumber)
	}
	return c.ReadCVFromFlash(index, cvNumber)
}

// Set sets the value of a CV, marking it as dirty if it's not read-only
func (c *Store) Set(cvNumber uint16, value uint8) bool {
	data, ok := c.data[cvNumber]
	if (data.Flags & ReadOnly) != 0 {
		return false // CV is read-only
	}
	if !ok || data.Value != value { // Only update if the value is different or key is missing
		data.Value = value
		data.Flags |= Dirty     // Mark as dirty
		c.data[cvNumber] = data // Store back into the map
	}

	return true
}

// Reset resets a single CV to its default value
func (c *Store) Reset(cvNumber uint16) bool {
	data, ok := c.data[cvNumber]
	if !ok {
		return false // CV not found or unused
	}
	if data.Value != data.Default { // Only update and mark dirty if needed
		data.Value = data.Default
		data.Flags |= Dirty
		c.data[cvNumber] = data
	}
	return true
}

// ResetAllCVs resets all used CVs to their default values
func (c *Store) ResetAll() {
	for cvNumber := range c.data {
		c.Reset(cvNumber)
	}
}

// ProcessChanges iterates through the CV table and processes any dirty CVs
// FIXME: Implement an index of handlers for each CV?
func (c *Store) ProcessChanges() {
	for cvNumber, data := range c.data {
		if (data.Flags & Dirty) != 0 {
			// Process the change based on the CV number
			switch cvNumber {
			// case 1: // Direction
			// 	SetMotorDirection(data.Value)
			// case 3: // Acceleration
			// 	SetAccelerationRate(data.Value)
			// case 4: // Deceleration
			// 	SetDecelerationRate(data.Value)
			// case 29: // Configuration Variable 1
			// 	UpdateConfiguration(data.Value)
			// ... handle other CVs ...
			default:
				// Handle unknown or unsupported changed CVs
			}

			// Clear the dirty flag
			data.Flags &^= Dirty    // Use bitwise AND NOT to clear the flag
			c.data[cvNumber] = data // Store back into the map

			// If persistent, save to flash
			if (data.Flags & Persistent) != 0 {
				c.Persist(cvNumber, data.Value)
			}
		}
	}
}

// Persist sets a CV value and immediately writes it to onboard flash
func (c *Store) Persist(cvNumber uint16, value uint8) bool {
	// Since this is for the current index, also set the current value
	c.Set(cvNumber, value)
	return c.persist(c.index, cvNumber, value)
}

// IndexedPersist writes a CV value to onboard flash using the provided index
func (c *Store) IndexedPersist(index, cvNumber uint16, value uint8) bool {
	return c.persist(index, cvNumber, value)
}
