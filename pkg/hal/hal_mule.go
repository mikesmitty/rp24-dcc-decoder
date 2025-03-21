//go:build pico2

package hal

import "machine"

const (
	// PWM frequency for the capacitor charge control pin
	CapChargeFreq = 1 * machine.MHz
	// PWM duty cycle for the capacitor charge control pin
	CapChargeDuty = 0.1

	// PWM frequency for the motor driver pins
	MaxMotorPWMFreq = 250 * machine.KHz
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

func (h *HAL) Init() {
	h.pins = make(map[string]machine.Pin)

	// MOSFET-backed function pins
	h.pins["aux1"] = machine.GPIO8
	h.pins["aux2"] = machine.GPIO5
	h.pins["aux5"] = machine.GPIO2
	h.pins["aux6"] = machine.GPIO9
	h.pins["aux7"] = machine.GPIO3
	h.pins["aux8"] = machine.GPIO4
	h.pins["lampFront"] = machine.GPIO6
	h.pins["lampRear"] = machine.GPIO7

	// GPIO-backed function pins
	h.pins["aux3"] = machine.GPIO14
	h.pins["aux4"] = machine.GPIO13
	h.pins["aux10"] = machine.GPIO10
	h.pins["aux11"] = machine.GPIO0
	h.pins["aux12"] = machine.GPIO1

	// Capacitor charge control pin (PWM at very low duty cycle)
	h.pins["capCharge"] = machine.GPIO21

	// Motor driver pins
	h.pins["adcRef"] = machine.GPIO28
	h.pins["backEMFA"] = machine.GPIO26
	h.pins["backEMFB"] = machine.GPIO27
	h.pins["motorA"] = machine.GPIO16
	h.pins["motorB"] = machine.GPIO17

	// Misc pins
	h.pins["led"] = machine.GPIO25

	// DCC pins
	h.pins["dcc"] = machine.GPIO22
	h.pins["railcom"] = machine.GPIO11

	// i2s audio out pins
	h.pins["i2sDIN"] = machine.GPIO18
	h.pins["i2sBCLK"] = machine.GPIO19
	h.pins["i2sLRCLK"] = machine.GPIO20

	// Initialize all pins
	for name, pin := range h.pins {
		switch name {
		case "dcc":
			// Enable internal pull-up since an N-channel MOSFET pulls GPIO21 low when rail polarity goes negative
			pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
		case "railcom":
			// RailCom logic levels are inverted, logic high is the low power state
			pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
			pin.High()
		default:
			// Set all pins to output low by default. This is the default GPIO state on reset for RP2xxx chips
			pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
			//pin.Low() // FIXME: this is causing the pins to be reset to low after being configured elsewhere
		}
	}
}

func (h *HAL) Pin(name string) (machine.Pin, bool) {
	_, ok := h.pins[name]
	return h.pins[name], ok
}
