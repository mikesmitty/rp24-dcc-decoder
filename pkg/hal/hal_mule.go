//go:build mule

package hal

import "machine"

func (h *HAL) Init() {
	clear(h.pins)

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
	h.pins["emfA"] = machine.GPIO26
	h.pins["emfB"] = machine.GPIO27
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
		case "dcc", "capCharge":
			// Enable internal pull-up since an N-channel MOSFET pulls GPIO21 low when rail polarity goes negative
			pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
		case "railcom", "led":
			// RailCom logic levels are inverted, logic high is the low power state
			pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
			pin.High()
		default:
			// Set all pins to output (low) by default. This is the default GPIO state on reset for RP2xxx chips
			pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		}
	}
}
