//go:build rp

package dcc

import (
	"time"
)

// bitPeriod is 4us +/- 2% (250kHz). We're using 3.5us to allow for compute overhead
const (
	bitPeriod = 3500 * time.Nanosecond

	nackByte byte = 0b00001111
	ackByte  byte = 0b11110000
	busyByte byte = 0b11100001
)

// From NMRA S-9.3.2
var rcByteCodes = [64]byte{
	0x00: 0b10101100, 0x01: 0b10101010, 0x02: 0b10101001, 0x03: 0b10100101,
	0x04: 0b10100011, 0x05: 0b10100110, 0x06: 0b10011100, 0x07: 0b10011010,
	0x08: 0b10011001, 0x09: 0b10010101, 0x0A: 0b10010011, 0x0B: 0b10010110,
	0x0C: 0b10001110, 0x0D: 0b10001101, 0x0E: 0b10001011, 0x0F: 0b10110001,
	0x10: 0b10110010, 0x11: 0b10110100, 0x12: 0b10111000, 0x13: 0b01110100,
	0x14: 0b01110010, 0x15: 0b01101100, 0x16: 0b01101010, 0x17: 0b01101001,
	0x18: 0b01100101, 0x19: 0b01100011, 0x1A: 0b01100110, 0x1B: 0b01011100,
	0x1C: 0b01011010, 0x1D: 0b01011001, 0x1E: 0b01010101, 0x1F: 0b01010011,
	0x20: 0b01010110, 0x21: 0b01001110, 0x22: 0b01001101, 0x23: 0b01001011,
	0x24: 0b01000111, 0x25: 0b01110001, 0x26: 0b11101000, 0x27: 0b11100100,
	0x28: 0b11100010, 0x29: 0b11010001, 0x2A: 0b11001001, 0x2B: 0b11000101,
	0x2C: 0b11011000, 0x2D: 0b11010100, 0x2E: 0b11010010, 0x2F: 0b11001010,
	0x30: 0b11000110, 0x31: 0b11001100, 0x32: 0b01111000, 0x33: 0b00010111,
	0x34: 0b00011011, 0x35: 0b00011101, 0x36: 0b00011110, 0x37: 0b00101110,
	0x38: 0b00110110, 0x39: 0b00111010, 0x3A: 0b00100111, 0x3B: 0b00101011,
	0x3C: 0b00101101, 0x3D: 0b00110101, 0x3E: 0b00111001, 0x3F: 0b00110011,
}

/*
func sendByte(pin machine.Pin, b byte) {
	// Start bit (zero bit, hardware logic is inverted)
	pin.High()
	/* Bit rate 250kHz +/- 2% (4us +/- 0.08us)
	We're using 3.5us here to allow for compute overhead and make up for lost time due to not
	having an accurate record of time since the last half-wave. The delay between the last
	half-wave and power cutout is 26-32us, but the timing is all keyed off that half-wave.
	* /
	time.Sleep(bitPeriod)

	// Data bits, LSB first
	for range 8 {
		if b&0x01 == 1 {
			pin.Low()
		} else {
			pin.High()
		}
		time.Sleep(bitPeriod)
		b >>= 1
	}

	// Stop bit (one bit, hardware logic is inverted)
	pin.Low()
	time.Sleep(bitPeriod)
}

var irqTriggered = false // FIXME: Cleanup

func (d *Decoder) AdvancedAck(ch1, ch2 []byte) error {
	if d.rcTxQueued {
		// Already queued, ignore
		return errors.New("AdvancedAck already queued")
	} else {
		fmt.Printf("AdvancedAck ch1: % X ch2: % X\r\n", ch1, ch2) // FIXME: Cleanup
	}
	d.rcTxQueued = true

	// When the capacitor charge pin is forced low we know the input voltage is below 8V,
	// and we use this as a proxy for the start of a RailCom cutout
	rcTxPin := d.rcTxPin.(machine.Pin)
	// d.capPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup}) // FIXME: Cleanup
	// err := d.capPin.SetInterrupt(machine.PinFalling, func(_ machine.Pin) { // FIXME: Allow for swapping logic?
	err := d.capPin.SetInterrupt(machine.PinRising, func(_ machine.Pin) {
		irqTriggered = true // FIXME: Cleanup
		now := time.Now()

		// Wait 54us to start squawking in channel 1
		time.Sleep(54 * time.Microsecond)

		// Send up to 2 bytes during channel 1
		for i := range ch1 {
			if i == 2 {
				break
			}
			sendByte(rcTxPin, ch1[i])
		}

		// Set the pin back to high (idle) and wait for channel 2 to start
		rcTxPin.High()
		time.Sleep(time.Until(now.Add(167 * time.Microsecond)))

		// Send up to 6 bytes during channel 2
		for i := range ch2 {
			if i == 6 {
				break
			}
			sendByte(rcTxPin, ch1[i])
		}

		// Set the pin back to high (idle)
		rcTxPin.High()
		d.rcTxQueued = false
	})
	if err != nil {
		// fmt.Printf("Error setting interrupt: %s\r\n", err) // FIXME: Cleanup
		return err
	}
	return nil
}
*/

func (d *Decoder) BasicAck() {
	// Pull all the power we can
	d.motor.ApplyPWM(1.0)
	for i := range d.outputPins {
		d.outputPins[i].High()
	}

	// Wait 6 +/- 1 ms
	time.Sleep(4 * time.Millisecond)

	// Turn everything off again
	d.motor.ApplyPWM(0.0)
	for i := range d.outputPins {
		d.outputPins[i].Low()
	}

}
