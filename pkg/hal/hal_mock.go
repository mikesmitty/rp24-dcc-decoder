//go:build !rp

package hal

type HAL struct {
	pins map[string]Pin
}

type Pin interface {
}

// Stub for non-RP platforms
func NewHAL() *HAL {
	return &HAL{}
}

// initI2SPIO is a stub for non-RP platforms
func (h *HAL) initI2SPIO(_ int, _, _ Pin) (I2S, error) {
	return nil, nil
}
