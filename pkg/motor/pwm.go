package motor

import (
	"math/rand/v2"

	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/hal"
)

func (m *Motor) initPWM(hw *hal.HAL, pinA, pinB shared.Pin, freq uint64, duty float32) error {
	if freq == 0 {
		freq = 40 * shared.KHz
		println("got pwm freq 0, using default motor PWM frequency of 40kHz")
	}

	pwmA, err := hw.InitPWM(pinA, freq, 0.0)
	if err != nil {
		return err
	}
	pwmB, err := hw.InitPWM(pinB, freq, 0.0)
	if err != nil {
		panic(err.Error())
	}
	m.pwmA = pwmA
	m.pwmB = pwmB

	return nil
}

// applyPWM sets the PWM outputs according to direction and duty cycle
func (m *Motor) applyPWM(dutyCycle float32) {
	// Take into account both m.reverse and m.ndotReverse to select a direction
	if m.Direction() == Reverse {
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
	m.pwmB.SetFreq(freq)
}
