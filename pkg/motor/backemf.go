package motor

import (
	"runtime"
	"time"

	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/hal"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/iir"
)

func (m *Motor) initBackEMF(emfA, emfB shared.Pin) {
	m.emfA = emfA
	m.emfB = emfB
	m.emfTicker = time.NewTicker(100 * time.Microsecond)
	m.emfTimer = time.NewTimer(100 * time.Microsecond)

	// Set up back EMF pins for ADC
	m.iir = iir.NewIIRFilter(m.iirAlpha)
	m.setupADC(m.Direction())
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
		if !m.DisablePID {
			m.emfValue = m.measureBackEMF()
		}
	}
}

// Measure the back EMF voltage to determine the motor speed
func (m *Motor) measureBackEMF() float32 {
	// Temporarily stop PWM for back EMF measurement
	m.stopMotor()

	// Lock the motor mutex to prevent concurrent access to the motor state
	m.pwmMutex.Lock()

	// Wait for EMF to settle
	time.Sleep(m.emfSettle)

	// Read back EMF every 100us during the cutout window
	m.emfTicker.Reset(100 * time.Microsecond)
	m.emfTimer.Reset(m.emfDuration)
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
	m.pwmMutex.Unlock()
	m.ApplyPWM(m.pwmDuty)

	return m.iir.Output()
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
