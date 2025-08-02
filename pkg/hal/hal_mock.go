//go:build !rp

package hal

import (
	"time"

	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
)

type HAL struct {
	pins map[string]shared.Pin
}

// Stub for non-RP platforms
func NewHAL() *HAL {
	return &HAL{}
}

// initI2SPIO is a stub for non-RP platforms
func (h *HAL) initI2SPIO(_ int, _, _ shared.Pin) (shared.I2S, error) {
	return nil, nil
}

func (h *HAL) InitPWM(pin shared.Pin, freq uint64, duty float32) (*SimplePWM, error) {
	return nil, nil
}

func (h *HAL) WatchdogSet(timeout time.Duration) {}

func (h *HAL) WatchdogReset() {}

type ADC struct{}

func NewADC(pin shared.Pin) *ADC {
	return &ADC{}
}

func (a *ADC) Read() uint16 {
	return 0
}
