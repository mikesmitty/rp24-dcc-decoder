//go:build rp

package hal

import (
	"machine"
	"time"
)

const (
	// PWM frequency for the capacitor charge control pin
	CapChargeFreq = 1 * machine.MHz
	// PWM duty cycle for the capacitor charge control pin
	CapChargeDuty = 0.1

	// PWM frequency for the motor driver pins
	DefaultMotorPWMFreq = 40 * machine.KHz
	MaxMotorPWMFreq     = 250 * machine.KHz
)

type HAL struct {
	pins map[string]machine.Pin
	pwms map[machine.Pin]uint8

	capChargeReady bool
}

func NewHAL() *HAL {
	h := &HAL{
		pins: make(map[string]machine.Pin),
		pwms: make(map[machine.Pin]uint8),
	}

	h.Init()

	return h
}

func (h *HAL) WatchdogSet(timeout time.Duration) {
	config := machine.WatchdogConfig{
		TimeoutMillis: uint32(timeout.Milliseconds()),
	}
	machine.Watchdog.Configure(config)

	machine.Watchdog.Start()
}

func (h *HAL) WatchdogReset() {
	machine.Watchdog.Update()
}
