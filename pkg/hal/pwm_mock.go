//go:build !rp

package hal

import "github.com/mikesmitty/rp24-dcc-decoder/internal/shared"

func (s *SimplePWM) Enable(enable bool) {
}

type PWM interface {
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
	Enable(bool)
	Top() uint32
	Configure(config shared.PWMConfig) error
	Channel(shared.Pin) (uint8, error)
}
