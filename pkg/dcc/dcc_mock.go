//go:build !rp

package dcc

import "github.com/mikesmitty/rp24-dcc-decoder/internal/shared"

// initPIO is a stub for non-RP platforms
func (d *Decoder) initPIO(pioNum int, pin shared.Pin) error {
	return nil
}
