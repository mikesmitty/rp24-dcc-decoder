package hal

import "machine"

// package machine
type PWM interface {
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
	Enable(bool)
	Top() uint32
	Configure(config machine.PWMConfig) error
	Channel(machine.Pin) (uint8, error)
}

type SimplePWM struct {
	channel uint8
	pwm     PWM
	slice   uint8
	top     float32
}
