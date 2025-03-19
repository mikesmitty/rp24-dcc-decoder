//go:build !rp

package dcc

// initPIO is a stub for non-RP platforms
func initPIO(pioNum int, pin Pin) error {
	return nil
}

func (d *Decoder) Reset() {}

func (d *Decoder) SetOpMode(mode opMode) {}
