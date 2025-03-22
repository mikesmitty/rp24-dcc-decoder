//go:build !rp

package dcc

// initPIO is a stub for non-RP platforms
func (d *Decoder) initPIO(pioNum int, pin Pin) error {
	return nil
}
