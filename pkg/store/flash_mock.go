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
