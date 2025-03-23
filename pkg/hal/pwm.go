//go:build rp

package hal

import (
	"machine"
)

type PWM interface {
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
	Enable(bool)
	Top() uint32
	Configure(config machine.PWMConfig) error
	Channel(machine.Pin) (uint8, error)
}

func (s *SimplePWM) Enable(enable bool) {
	s.pwm.Enable(enable)
}

func (s *SimplePWM) SetDuty(duty float32) {
	s.pwm.Set(s.channel, uint32(duty*s.top))
}

func (s *SimplePWM) SetFreq(freq uint64) {
	s.pwm.SetPeriod(1e9 / freq)
}

func (s *SimplePWM) Slice() uint8 {
	return s.slice
}

var pwms = [...]PWM{machine.PWM0, machine.PWM1, machine.PWM2, machine.PWM3, machine.PWM4, machine.PWM5, machine.PWM6, machine.PWM7}

func (h *HAL) InitPWM(pin machine.Pin, freq uint64, duty float32) (*SimplePWM, error) {
	slice, err := machine.PWMPeripheral(pin)
	if err != nil {
		return nil, err
	}

	pwm := pwms[slice]

	channel, err := pwm.Channel(pin)
	if err != nil {
		return nil, err
	}

	err = pwm.Configure(machine.PWMConfig{Period: 1e9 / freq})
	if err != nil {
		return nil, err
	}

	spwm := &SimplePWM{
		channel: channel,
		pwm:     pwm,
		slice:   slice,
		top:     float32(pwm.Top()),
	}

	spwm.SetDuty(duty)
	spwm.Enable(true)

	return spwm, nil
}
