package motor

import (
	"runtime"
	"time"

	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/hal"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/iir"
)

const (
	emfSettleWait = 100 * time.Microsecond // TODO: Tune these values
	emfReadDelay  = 100 * time.Microsecond
)

func (m *Motor) initBackEMF(emfA, emfB, adcRef shared.Pin) {
	m.emfA = emfA
	m.emfB = emfB
	m.emfTicker = time.NewTicker(emfSettleWait)
	m.emfTimer = time.NewTimer(emfSettleWait)

	// Set up back EMF pins for ADC
	m.iir = iir.NewIIRFilter(m.iirAlpha)
	m.setupADC(m.Direction())
	if adcRef != shared.NoPin {
		m.adcRef = hal.NewADC(adcRef)
		m.iirRef = iir.NewIIRFilter(m.iirAlpha)
	}
}

func (m *Motor) setupADC(direction Direction) {
	// Set up back EMF pin for ADC
	pin := m.emfA
	if direction == Reverse {
		pin = m.emfB
	}
	m.iir.Reset()
	m.emfADC = hal.NewADC(pin)
}

func (m *Motor) RunEMF() {
	for {
		time.Sleep(m.emfInterval)
		m.emfValue = m.measureBackEMF()
	}
}

// Measure the back EMF voltage to determine the motor speed
func (m *Motor) measureBackEMF() float32 {
	// Lock the motor mutex to prevent concurrent access to the motor state
	m.emfMutex.Lock()
	defer m.emfMutex.Unlock()

	// Temporarily stop PWM for back EMF measurement
	m.stopMotor()

	// Wait for motor to stop
	time.Sleep(100 * time.Millisecond)

	// Kill time waiting for the motor to stop by measuring the ADC reference
	m.emfTicker.Reset(20 * time.Microsecond)
	m.emfTimer.Reset(emfSettleWait)
PREP:
	for {
		select {
		// Wait for the back EMF to settle before measuring
		case <-m.emfTimer.C:
			// EMF settling time is done, break out of the double loop
			m.emfTicker.Stop()
			break PREP
		case <-m.emfTicker.C:
			// Measure the ADC reference attached to ground to help establish a noise baseline while we wait
			m.iirRef.Filter(float32(m.adcRef.Read()))
		default:
			// Allow other tasks to run while waiting
			runtime.Gosched()
		}
	}

	m.emfTicker.Reset(emfReadDelay)
	m.emfTimer.Reset(m.emfDuration - emfSettleWait)
DONE:
	for {
		select {
		// Wait until the end of the PWM cutout period (excluding the already-completed motor settling time)
		case <-m.emfTimer.C:
			// Cutout period is over, break out of the double loop
			m.emfTicker.Stop()
			break DONE
		case <-m.emfTicker.C:
			// Measure back EMF into the IIR filter and sleep until the next one
			m.iir.Filter(float32(m.emfADC.Read()))
		default:
			// Allow other tasks to run while waiting
			runtime.Gosched()
		}
	}

	// Restore PWM
	m.applyPWM(m.pwmDuty)

	return m.iir.Output() - m.iirRef.Output()
}

func (m *Motor) updateBackEMFTiming() {
	// Calculate the cutout interval and duration in 100us units based on the current speed setting
	m.emfInterval = m.varyBySpeed(m.cv[116], m.cv[117])
	m.emfDuration = m.varyBySpeed(m.cv[118], m.cv[119])
}

// Calculate a varying offset based on the currently commanded speed
func (m *Motor) varyBySpeed(min, max uint8) time.Duration {
	// Calculate the spread of the speed table
	// This is used to calculate the back EMF interval by speed
	step := float32(max-min) / float32(m.speedMode)
	count := time.Duration(float32(min) + float32(m.currentSpeed)*step)
	return count * 100 * time.Microsecond
}
