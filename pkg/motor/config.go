//go:build rp

package motor

import (
	"machine"
	"time"

	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
)

func (m *Motor) CVCallback() shared.CVCallbackFunc {
	return func(cvNumber uint16, value uint8) bool {
		switch cvNumber {
		case 2, 5, 6:
			// CV2 Vstart (minimum throttle required to start moving)
			// CV5 Vmax (max throttle)
			// CV6 Vmid (mid throttle)
			defer m.updateSpeedTable()

		case 3, 4, 23, 24:
			// CV3 Acceleration rate
			// CV4 Deceleration rate
			// CV23 Consist acceleration modifier
			// CV24 Consist deceleration modifier
			defer m.calculateAccelDecelRates()

		case 9:
			// PWM freq in kHz (1-250)
			value = max(1, min(value, 250))
			// Update PWM frequency
			m.setPWMFreq(uint64(value) * machine.KHz)

		case 10:
			// Back EMF motor control cutoff speed
			// TODO: Implement this

		case 19:
			// Consist address direction swap modifier
			// Check for double-uno-reverse
			m.ndotReverse = (value >> 7) != (m.cv[29] & 1)

		case 29:
			// CV 29:
			// Bits 7-5 are not relevant here
			// Ignore bit 3, not concerned about RailCom here
			// Ignore bit 2, not concerned about DC mode
			if (value & 0b00000010) == 0 {
				m.speedMode = SpeedMode14
			} else {
				// Default to 28-speed mode, 128-speed mode is enabled automatically
				// when a 128-speed packet is received
				m.speedMode = SpeedMode28
			}
			// ndotReverse uno-reverses the normal direction of travel
			m.ndotReverse = (value & 1) != (m.cv[19] >> 7)

			// Update speed table in case bit 4 or bit 1 changed
			// Bit 4: 0 = CV 2,5,6 speed curve, 1 = CV 67-94 speed table
			defer m.updateSpeedTable()

		case 53:
			// Max speed EMF voltage
			m.emfMax = float32(value) / 100

		case 51, 52, 54, 55, 56:
			// CV51 Kp gain cutover speed step
			// CV52 Low speed Kp gain (proportional)
			// CV54 High speed Kp gain (proportional)
			// CV55 Ki gain (integral)
			// CV56 Low speed PID scaling factor
			defer m.updatePIDConfig() // TODO: Is this the right idea? Not sure how to handle high/low speed switch

		case 65:
			// Startup kick to overcome static friction from a stop to speed step 1
			// TODO: Implement this

		case 66, 95:
			// Forward/reverse trim - n/128 * throttle vs. the opposite direction
			// e.g. 64/128 reduces throttle by half, 192/128 increases throttle by 50%
			defer m.updateDirectionTrims()

		case 116, 117:
			// Back EMF measurement interval in 100us steps (50-200) 5-20ms
			value = max(50, min(value, 200))
			defer m.updateBackEMFTiming()

		case 118, 119:
			// Back EMF measurement duration in 100us steps (10-40) 1-4ms
			value = max(10, min(value, 40))
			defer m.updateBackEMFTiming()
		}

		// Special case for the 28-CV speed table
		if cvNumber >= 67 && cvNumber <= 94 {
			// Updates to the user speed table
			defer m.updateSpeedTable()
		}

		// Update the cached CV value
		m.cv[cvNumber] = value
		return true
	}
}

func (m *Motor) AckPulse() {
	// FIXME: Implement this properly?
	dutyMax := float32(1.0)

	// Full steam ahead 3ms, full reverse 3ms, then stop
	m.pwmA.SetDuty(dutyMax)
	m.pwmB.SetDuty(0.0)
	time.Sleep(3 * time.Millisecond)

	m.pwmA.SetDuty(0.0)
	m.pwmB.SetDuty(dutyMax)
	time.Sleep(3 * time.Millisecond)

	m.pwmA.SetDuty(0.0)
	m.pwmB.SetDuty(0.0)
}

// calculateAccelDecelRates updates the acceleration and deceleration rates
// based on CVs 3, 4, 23, and 24
func (m *Motor) calculateAccelDecelRates() {
	// Get base values from CV3 and CV4
	accelBase := float32(m.cv[3])
	decelBase := float32(m.cv[4])

	// Get adjustment values from CV23 and CV24
	accelAdj := float32(m.cv[23] & 0x7F) // Lower 7 bits for magnitude
	decelAdj := float32(m.cv[24] & 0x7F) // Lower 7 bits for magnitude

	// Apply sign based on bit 7
	if (m.cv[23] & 0x80) != 0 {
		accelAdj = -accelAdj
	}
	if (m.cv[24] & 0x80) != 0 {
		decelAdj = -decelAdj
	}

	// Calculate acceleration rate with formula:
	// (CV3 + adjustment from CV23) * 0.896 / number of speed steps
	m.accelRate = max(0, (accelBase+accelAdj)*0.896/float32(m.speedMode))

	// Calculate deceleration rate with formula:
	// (CV4 + adjustment from CV24) * 0.896 / number of speed steps
	m.decelRate = max(0, (decelBase+decelAdj)*0.896/float32(m.speedMode))
}

// updatePIDConfig updates the PID controller based on CVs 51, 52, 54, 55, and 56
func (m *Motor) updatePIDConfig() {
	// Get the cutover speed step
	m.kpCutover = m.cv[51]

	// Get the low speed Kp gain
	m.kpLow = float32(m.cv[52]) / 10

	// Get the high speed Kp gain
	m.kpHigh = float32(m.cv[54]) / 10

	// Update the Kp gain based on the current speed
	if m.currentSpeed < m.kpCutover {
		// Use the low speed Kp gain
		m.pid.Config.ProportionalGain = m.kpLow
	} else {
		// Use the high speed Kp gain
		m.pid.Config.ProportionalGain = m.kpHigh
	}

	// Get the Ki gain
	m.pid.Config.IntegralGain = float32(m.cv[55]) / 10
}

// TODO: Make sure to update the backemf interval when speed changes
// updateSpeedTable generates the speed table based on CV67-94 and other settings
func (m *Motor) updateSpeedTable() {
	if m.useSpeedTable {
		// Generate the speed table based on CV67-94
		m.generateUserSpeedTable()
	} else {
		// Generate the speed table based on CV2, 5, and 6
		m.generate3PointSpeedTable()
	}
}

// generate3PointSpeedTable creates a speed table using Vstart, Vmid, and Vmax
func (m *Motor) generate3PointSpeedTable() {
	vStart := float32(m.cv[2])
	vMax := float32(m.cv[5])
	if vMax == 0 {
		vMax = 255
	}
	vMid := float32(m.cv[6])
	if vMid == 0 {
		vMid = (vStart + vMax) / 2
	}

	// Clear the table
	for i := range m.speedTable {
		m.speedTable[i] = 0
	}

	// Get number of non-stationary speed steps and interpolate
	steps := float32(m.speedMode)
	for i := float32(0); i < steps; i++ {
		var value float32
		if i <= vMid {
			// First segment (between Vstart and Vmid)
			ratio := (i + 1) / vMid
			value = vStart + ratio*(vMid-vStart)
		} else {
			// Second segment (between Vmid and Vmax)
			ratio := (i + 1 - vMid) / (steps - 1 - vMid)
			value = vMid + ratio*(vMax-vMid)
		}
		m.speedTable[int(i)+2] = float32(value) / 255
	}
}

// generateUserSpeedTable creates a speed table based on CV67-94, interpolating to 128 steps if necessary
func (m *Motor) generateUserSpeedTable() {
	// Clear the table
	for i := range m.speedTable {
		m.speedTable[i] = 0
	}

	switch m.speedMode {
	case SpeedMode14:
		// Skip every other CV value in 14-speed mode
		for i := uint16(0); i < 14; i++ {
			// Use every other CV value
			m.speedTable[i+2] = float32(m.cv[2*i+67]) / 255
		}
		m.speedTable[15] = float32(m.cv[94]) / 255
	case SpeedMode28:
		// Use all CV values
		for i := uint16(0); i < 28; i++ {
			m.speedTable[i+2] = float32(m.cv[i+67]) / 255
		}
	case SpeedMode128:
		// Interpolate to 128 steps
		for i := uint16(0); i < 128; i++ {
			// Calculate the index in the 28-speed table
			index := i * 28 / 128
			// Interpolate between the two values
			value := float32(m.cv[index+67]) + float32(m.cv[index+68]-m.cv[index+67])*float32(i)*28/128
			m.speedTable[i+2] = value / 255
		}
	default:
		panic("Invalid speed mode")
	}
}

// updateDirectionTrims updates the forward and reverse trim values
func (m *Motor) updateDirectionTrims() {
	// Make sure we can't zero out the trims and prevent throttle from working at all
	fwd := m.cv[66]
	if fwd == 0 {
		fwd = 128
	}
	rev := m.cv[95]
	if rev == 0 {
		rev = 128
	}
	// Forward trim
	m.fwdTrim = float32(fwd) / 128
	// Reverse trim
	m.revTrim = float32(rev) / 128
}
