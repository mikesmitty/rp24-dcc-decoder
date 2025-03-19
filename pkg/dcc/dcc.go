//go:build rp

package dcc

import (
	"errors"
	"machine"
	"time"

	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
)

//go:generate pioasm -o go dcc.pio dcc_pio.go

func NewDecoder(cvHandler cv.Handler, pioNum int, pin machine.Pin) (*Decoder, error) {
	d := &Decoder{address: make([]byte, 2), cv: cvHandler}

	err := d.initPIO(pioNum, pin)
	if err != nil {
		return nil, err
	}

	// FIXME: Cleanup
	// d.address[0] = 0xC0
	// d.address[1] = 150
	d.SetAddress(150) // FIXME: Testing

	return d, nil
}

// FIXME: Cleanup?
// Set the address of the DCC reader
func (d *Decoder) SetAddress(addr uint16) error {
	if addr == 0 || addr > 10239 {
		return errors.New("address out of range")
	}

	d.address = d.address[:0]
	if addr > 127 {
		d.address = append(d.address, 0xC0|byte(addr>>8))
		d.address = append(d.address, byte(addr))
	} else {
		d.address = append(d.address, byte(addr))
	}

	return nil
}

// FIXME: Cleanup?
func (d *Decoder) Address() []byte {
	return d.address
}

// Enable or disable the DCC reader
func (d *Decoder) Enable(enabled bool) {
	d.sm.SetEnabled(enabled)
}

func (d *Decoder) OpMode() opMode {
	return d.opMode
}

func (d *Decoder) SetOpMode(mode opMode) {
	if mode == ServiceMode {
		// If we've received a service mode reset packet re-up the 20ms timer
		d.lastSvcResetTime = time.Now()
		d.svcModeReady = true
	} else {
		d.svcModeReady = false
	}
	d.opMode = mode
	// FIXME: Handle entering/exiting service mode
}

func (d *Decoder) Reset() {
	d.lastSvcResetTime = time.Now()
	d.svcModeReady = true
	// FIXME: Implement reset packet handling
	/* When a Digital Decoder receives a Digital Decoder Reset Packet, it shall erase all
	volatile memory (including any speed and direction data), and return to its normal
	power-up state. If the Digital Decoder is operating a locomotive at a non-zero speed
	when it receives a Digital Decoder Reset, it shall bring the locomotive to an
	immediate stop.  */
}

func (d *Decoder) CVCallback() cv.CVCallbackFunc {
	return func(cvNumber uint16, value uint8) bool {
		switch cvNumber {
		case 1:
			// Set the short address
			// Not allowing 0 because DC mode is out of scope
			if value >= 1 && value <= 127 {
				d.SetAddress(uint16(value))
			} else {
				return false
			}
		case 17, 18:
			// CV17 must be in the range 192-231, CV18 can be any value
			// The top 2 bits of CV17 are ignored when parsing the address
			if cvNumber == 17 && (value < 192 || value > 231) {
				return false
			}
			// Set the extended address bytes
			d.address[cvNumber-17] = value
		case 19, 20:
			// Set the consist address bytes
			// FIXME: Implement, including reverse ndot
		case 29:
			// CV29 bit 5: 0 = short address, 1 = extended address
			cv17 := d.cv.CV(17)
			if (value & 0b00100000) != 0 {
				if cv17 >= 192 && cv17 <= 231 {
					// If CV17 is 192 and CV18 is 0 the long address would be 0, abort
					if cv17 == 192 && d.cv.CV(18) == 0 {
						return false
					}
					// Otherwise, if CV17 is valid, set bit 5 and enable it
					d.address = append(d.address[:0], 0xC0|d.cv.CV(17))
					d.address = append(d.address, d.cv.CV(18))
				} else {
					// If CV17 is invalid, reject the CV29 update
					return false
				}
			} else {
				// Clear bit 5 and use the short address
				d.address = append(d.address[:0], d.cv.CV(1))
				return false
			}
		}
		return true
	}
}

/* FIXME: Might be useful?
// SetSampleFrequency sets the sample frequency of the I2S peripheral.
func (d *DCCReader) SetFrequency(freq uint32) error {
	freq *= 32 // 32 bits per sample
	whole, frac, err := pio.ClkDivFromFrequency(freq, machine.CPUFrequency())
	if err != nil {
		return err
	}
	i2s.sm.SetClkDiv(whole, frac)
	return nil
}
*/
