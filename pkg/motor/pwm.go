//go:build rp

package motor

import (
	"machine"
	"math/rand/v2"

	"github.com/mikesmitty/rp24-dcc-decoder/pkg/hal"
)

func (m *Motor) initPWM(hw *hal.HAL, pinA, pinB machine.Pin, freq uint32, duty float32) error {
	pwmA, err := hw.InitPWM(pinA, hal.MaxMotorPWMFreq, 0.0)
	if err != nil {
		return err
	}
	pwmB, err := hw.InitPWM(pinB, hal.MaxMotorPWMFreq, 0.0)
	if err != nil {
		panic(err.Error())
	}
	m.pwmA = pwmA
	m.pwmB = pwmB

	if pwmA.Slice() == pwmB.Slice() {
		// If the pins are on the same PWM slice, they should share the same frequency
		// so we only need to set it once
		m.pwmShared = true
	}

	return nil
}

// applyPWM sets the PWM outputs according to direction and duty cycle
func (m *Motor) applyPWM(dutyCycle float32) {
	// One train conductor always drives in reverse, the other always sits backwards
	// If they both agree on a direction, go the other way
	if m.reverse == m.ndotReverse {
		m.pwmA.SetDuty(0.0)
		m.pwmB.SetDuty(dutyCycle * m.revTrim)
	} else {
		m.pwmA.SetDuty(dutyCycle * m.fwdTrim)
		m.pwmB.SetDuty(0.0)
	}
}

// Dither the motor PWM frequency to improve low-speed startup. Window size represents the
// maximum amount of dithering to apply in 100Hz steps
func (m *Motor) dither(freq, windowSize uint64) {
	modifier := rand.Uint64N(windowSize)
	neg := modifier & 1
	if neg == 1 {
		modifier = -modifier
	}
	modifier *= 100
	m.setPWMFreq(freq + modifier)
}

// setPWMFreq sets the PWM frequency for the motor driver signal
func (m *Motor) setPWMFreq(freq uint64) {
	m.pwmA.SetFreq(freq)
	if !m.pwmShared {
		m.pwmB.SetFreq(freq)
	}
}
