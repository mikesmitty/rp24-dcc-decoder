//go:build rp

package motor

import (
	"machine"
	"runtime"
	"time"

	"github.com/mikesmitty/rp24-dcc-decoder/pkg/hal"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/iir"
)

func (m *Motor) initBackEMF(emfA, emfB, adcRef machine.Pin) {
	// Set up back EMF pins for ADC
	m.emfA = hal.NewADC(emfA)
	m.iirA = iir.NewIIRFilter(m.iirAlpha)
	m.emfB = hal.NewADC(emfB)
	m.iirB = iir.NewIIRFilter(m.iirAlpha)
	if adcRef != machine.NoPin {
		m.adcRef = hal.NewADC(adcRef)
		m.iirRef = iir.NewIIRFilter(m.iirAlpha)
	}

	// Set up back EMF measurement interval
	m.emfTrigger = make(chan struct{})
}

// Calculate estimated speed based on back EMF
func (m *Motor) getEstimatedSpeed() float32 {
	return m.emfReading / m.emfMax
}

// Trigger the back EMF measurement interval dynamically based on current emfInterval
func (m *Motor) bemfInterval() {
	for {
		time.Sleep(m.emfInterval)
		// Announce that the back EMF measurement interval has elapsed and should be measured
		m.emfTrigger <- struct{}{}
	}
}

func (m *Motor) updateBackEMFTiming() {
	// Calculate the cutout interval and duration in 100us units based on the current speed setting
	m.emfInterval = time.Duration(m.varyBySpeed(m.cv[116], m.cv[117]))
	m.emfDuration = time.Duration(m.varyBySpeed(m.cv[118], m.cv[119]))
}

// FIXME: Implement this properly
// CalculateADCOffset takes n samples from each ADC channel and calculates the average
func (m *Motor) CalculateADCOffset(n int) {
	// Store current motor state
	wasRunning := m.currentSpeed > 0

	// Stop the motor
	m.stopMotor()

	// Wait for motor to stop
	time.Sleep(100 * time.Millisecond)

	// Take n samples from each ADC channel
	sumA := uint32(0)
	sumB := uint32(0)
	sumRef := uint32(0)

	for i := 0; i < n; i++ {
		sumA += uint32(m.emfA.Read())
		sumB += uint32(m.emfB.Read())
		if m.adcRef != nil {
			sumRef += uint32(m.adcRef.Read())
		}
		time.Sleep(time.Millisecond)
	}

	// Calculate average values
	// emfAAvg := float32(sumA) / float32(n)
	// emfBAvg := float32(sumB) / float32(n)

	// If reference ADC is available, use it for calibration
	if m.adcRef != nil {
		//adcRefAvg := float32(sumRef) / float32(n)
		// Apply calibration based on reference
		// This depends on your specific hardware design
	}

	// Restore motor state if it was running
	if wasRunning {
		m.SetSpeed(m.currentSpeed, !m.reverse)
	}
}

// Calculate a varying offset based on the currently commanded speed
func (m *Motor) varyBySpeed(min, max uint8) uint8 {
	// Calculate the spread of the speed table
	// This is used to calculate the back EMF interval by speed
	step := float32(max-min) / float32(m.speedMode)
	return min + uint8(float32(m.currentSpeed)*step)
}

// FIXME: Look this over
// measureBackEMF measures back EMF voltage from the motor
func (m *Motor) measureBackEMF() float32 {
	const settleWait = 100 * time.Microsecond
	const readDelay = 100 * time.Microsecond

	// Select the appropriate ADC channel based on direction
	var adc *hal.ADC
	var iir *iir.IIRFilter
	var offset float32

	refADC := m.adcRef
	refIIR := m.iirRef

	if m.reverse {
		adc = m.emfB
		iir = m.iirB
		offset = m.adcOffsetB
	} else {
		adc = m.emfA
		iir = m.iirA
		offset = m.adcOffsetA
	}

	// FIXME: We have the throttle setting, we shouldn't need the raw PWM values?
	// Temporarily stop PWM for back EMF measurement
	currentPWMA := float32(1) // m.pwmA.GetDuty()
	currentPWMB := float32(1) // m.pwmB.GetDuty()
	m.stopMotor()

PREP:
	for {
		select {
		// Wait for the back EMF to settle before measuring
		case <-time.After(settleWait):
			// EMF settling time is done, break out of the double loop
			break PREP
		case <-time.After(20 * time.Microsecond):
			// Measure the ADC reference attached to ground to help establish a noise baseline while we wait
			refIIR.Filter(float32(refADC.Read()))
		default:
			// Allow other tasks to run while waiting
			runtime.Gosched()
		}
	}

DONE:
	for {
		select {
		// Wait until the end of the PWM cutout period (excluding the already-completed motor settling time)
		case <-time.After(m.emfDuration - settleWait):
			// Cutout period is over, break out of the double loop
			break DONE
		case <-time.After(readDelay):
			// Measure back EMF into the IIR filter and sleep until the next one
			iir.Filter(float32(adc.Read()))
			time.Sleep(readDelay)
		default:
			// Allow other tasks to run while waiting
			runtime.Gosched()
		}
	}

	// FIXME: This logic is redundant and gross. Refactor it
	// Restore initial PWM
	m.applyPWM(currentPWMA)
	if m.reverse {
		m.pwmA.SetDuty(0.0)
		m.pwmB.SetDuty(currentPWMB)
	} else {
		m.pwmA.SetDuty(currentPWMA)
		m.pwmB.SetDuty(0.0)
	}

	return iir.Output() - offset
}
