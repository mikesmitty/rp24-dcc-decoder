//go:build !rp

package hal

import "github.com/mikesmitty/rp24-dcc-decoder/internal/shared"

var PWMSetDutyHook func(pwm *SimplePWM, duty float32)
var PWMSetFreqHook func(pwm *SimplePWM, freq uint64)

func (s *SimplePWM) Enable(enable bool) {
}

func (s *SimplePWM) SetDuty(duty float32) {
	if PWMSetDutyHook != nil {
		PWMSetDutyHook(s, duty)
	}
}

func (s *SimplePWM) SetFreq(freq uint64) {
	if PWMSetFreqHook != nil {
		PWMSetFreqHook(s, freq)
	}
}

type pwm interface {
	Set(channel uint8, value uint64)
	SetPeriod(period uint64) error
	Enable(bool)
	Top() uint32
	Configure(config shared.PWMConfig) error
	Channel(shared.Pin) (uint8, error)
}

