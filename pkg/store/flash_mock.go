//go:build !rp

package store

func (c *Store) initFlash() error {
	return nil
}

func (c *Store) persist(index, cvNumber uint16, value uint8) bool {
	return true
}

func (c *Store) ReadCVFromFlash(index, cvNumber uint16) (uint8, bool) {
	return 0, true
}

func (c *Store) ProcessChanges() bool {
	// Clear the dirty flags
	for cvNumber, data := range c.data {
		if (data.Flags & Dirty) != 0 {
			data.Flags &^= Dirty    // Use bitwise AND NOT to clear the flag
			c.data[cvNumber] = data // Store back into the map
		}
	}

	return true
}

// Bool return value indicates if the index file was found
func (s *Store) LoadIndex(newIndex uint16) (bool, error) {
	if newIndex > 0 {
		return false, nil
	}
	return true, nil
}
