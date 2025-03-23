package dcc

import (
	"errors"
	"time"

	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
)

//go:generate pioasm -o go dcc.pio dcc_pio.go

func NewDecoder(cvHandler cv.Handler, pioNum int, pin shared.Pin) (*Decoder, error) {
	d := &Decoder{
		address:         make([]byte, 0, 2),
		consistAddress:  make([]byte, 0, 2),
		cv:              cvHandler,
		outputCallbacks: make(map[uint16][]shared.OutputCallback, 12),
		outputMapsFwd:   make(map[uint16]uint16, 12),
		outputMapsRev:   make(map[uint16]uint16, 12),
	}

	err := d.initPIO(pioNum, pin)
	if err != nil {
		return nil, err
	}

	// TODO: Handle setting version number at build time

	d.RegisterCallbacks()

	return d, nil
}

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
	// TODO: Handle exiting service mode
}

func (d *Decoder) Reset() {
	d.lastSvcResetTime = time.Now()
	d.svcModeReady = true
	// TODO: Implement reset packet handling
	/* When a Digital Decoder receives a Digital Decoder Reset Packet, it shall erase all
	volatile memory (including any speed and direction data), and return to its normal
	power-up state. If the Digital Decoder is operating a locomotive at a non-zero speed
	when it receives a Digital Decoder Reset, it shall bring the locomotive to an
	immediate stop.  */
}

func (d *Decoder) RegisterCallbacks() {
	d.cv.RegisterCallback(1, d.CVCallback())
	for i := uint16(17); i <= 22; i++ {
		d.cv.RegisterCallback(i, d.CVCallback())
	}
	d.cv.RegisterCallback(29, d.CVCallback())
	for i := uint16(33); i <= 46; i++ {
		d.cv.RegisterCallback(i, d.CVCallback())
	}
}

func (d *Decoder) CVCallback() shared.CVCallbackFunc {
	return func(cvNumber uint16, value uint8) bool {
		switch cvNumber {
		case 1:
			// Set the short address
			// Not allowing 0 because DC mode is out of scope
			if value < 1 || value > 127 {
				return false
			}
			d.SetAddress(uint16(value))

		case 17, 18:
			// CV17 must be in the range 192-231, CV18 can be any value
			// The top 2 bits of CV17 are ignored when parsing the address
			if cvNumber == 17 && (value < 192 || value > 231) {
				return false
			}
			// Set the extended address bytes
			if len(d.address) < int(cvNumber)-16 {
				d.address = append(d.address, value)
			} else {
				d.address[cvNumber-17] = value
			}

		case 19, 20:
			// Set the consist address bytes
			// TODO: Add validation around CV19 value in extended address mode when standardized
			if len(d.address) < int(cvNumber)-18 {
				d.address = append(d.address, value)
			} else {
				d.address[cvNumber-19] = value
			}

		case 21:
			// Convert CV21 to a bitmask for enabling the functions via consist address (F1-F8)
			// Clear the bits for F1-F4
			mask := d.consistFuncMask[0] & 0b11110000
			// Set the bits for F1-F4
			mask |= value & 0b00001111
			d.consistFuncMask[0] = mask

			// Set the bits for F5-F8
			d.consistFuncMask[1] = value >> 4

		case 22:
			// Convert CV22 to a bitmask for enabling the functions via consist address (FLf, FLr, F9-F12)
			// TODO: Implement this for FLf and FLr separately
			// Clear the bit for FL (F0)
			mask := d.consistFuncMask[0] &^ (1 << 4)
			// Set the bit for FL (F0) (CV bit 0 -> mask bit 4, FLr is CV bit 1)
			mask |= (value & 1) << 4
			d.consistFuncMask[0] = mask

			// Set the mask bits for F9-F12 (CV bits 2-5)
			d.consistFuncMask[2] = (value & 0b00111100) >> 2

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
		case 33:
			// Configure function mapping to output F0f
			d.outputMapsFwd[0] = uint16(value)
		case 34:
			// Configure function mapping to output F0r
			d.outputMapsRev[0] = uint16(value)
		case 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46:
			// Configure function mapping for F1-F12 to outputs AUX1-AUX12 (fwd and rev)
			outputNum := cvNumber - 33
			outputs := uint16(value)
			if cvNumber >= 43 {
				// AUX5-AUX12
				outputs = outputs << 6
			} else if cvNumber >= 38 {
				// AUX2-AUX9
				outputs = outputs << 3
			}
			d.outputMapsFwd[outputNum] = uint16(value)
			d.outputMapsRev[outputNum] = uint16(value)
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
