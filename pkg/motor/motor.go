package motor

import (
	"fmt"
	"sync"
	"time"

	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/hal"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/iir"
	"github.com/mikesmitty/tinypid"
)

type Motor struct {
	cv        map[uint16]uint8
	cvHandler cv.Handler

	// H-bridge PWM outputs
	pwmA        *hal.SimplePWM
	pwmB        *hal.SimplePWM
	pwmDuty     float32
	pwmInterval time.Duration

	iirAlpha float32
	iir      *iir.IIRFilter
	iirRef   *iir.IIRFilter

	// Back-EMF measurement state for estimating speed
	adcRef      *hal.ADC
	emfADC      *hal.ADC
	emfA        shared.Pin
	emfB        shared.Pin
	emfDuration time.Duration
	emfInterval time.Duration
	emfMax      float32
	pwmMutex    sync.Mutex
	emfTarget   float32
	emfTicker   *time.Ticker
	emfTimer    *time.Timer
	emfValue    float32

	// Motor control state
	currentSpeed    uint8     // Current speed step/speed table index
	changeDirection bool      // Flag to indicate direction change commanded
	ndotReverse     bool      // Normal direction of travel is reversed
	reverse         bool      // Commanded direction of travel is reversed
	rawSpeed        float32   // currentSpeed without rounding for accel/decel rates
	speedAfterStop  uint8     // Next speed step after direction change
	speedMode       SpeedMode // Number of speed steps
	targetRaw       float32   // targetSpeed without rounding
	targetSpeed     uint8     // Commanded speed step

	accelRate float32 // Max steps per second accel
	decelRate float32 // Max steps per second decel
	fwdTrim   float32 // Forward speed trim
	revTrim   float32 // Reverse speed trim

	// Speed table for the max 128 speed steps (including stop and e-stop)
	speedTable     [128]float32
	userSpeedTable bool

	// For PID control
	DisablePID bool
	pid        tinypid.PIController
	kpCutover  uint8 // Speed step at which to switch from low to high PID gains
	kpLow      float32
	kpHigh     float32

	// For PID control
	lastControlTime time.Time
}

func NewMotor(conf cv.Handler, hw *hal.HAL, pinA, pinB, emfA, emfB, adcRef shared.Pin) *Motor {
	m := &Motor{
		cv:              make(map[uint16]uint8),
		cvHandler:       conf,
		iirAlpha:        0.7,
		lastControlTime: time.Now(),
	}

	// Set up back EMF pins for ADC
	m.initBackEMF(emfA, emfB, adcRef)

	// Set up motor driver pins for PWM
	m.pwmInterval = 100 * time.Millisecond // TODO: Adjust this?
	err := m.initPWM(hw, pinA, pinB, uint64(m.cvHandler.CV(9))*shared.KHz, 0.0)
	if err != nil {
		panic(err.Error())
	}

	m.RegisterCallbacks()

	return m
}

func (m *Motor) Run() {
	for {
		time.Sleep(m.pwmInterval)
		m.runMotorControl()
	}
}

// runMotorControl is the main control loop for the motor
func (m *Motor) runMotorControl() {
	now := time.Now()
	elapsed := now.Sub(m.lastControlTime)

	// Update speed step
	prevSpeed := m.currentSpeed
	m.updateSpeedStep(elapsed)
	if prevSpeed != m.currentSpeed {
		// If the speed step changed, update the PID controller config (update proportional gain for new speed step) and EMF target
		m.updatePIDConfig()
		m.emfTarget = m.speedTable[m.currentSpeed] * m.emfMax
	}

	// Run the PID loop
	m.pid.Update(tinypid.PIControllerInput{
		ReferenceSignal: m.emfTarget,
		ActualSignal:    m.emfValue,
	})

	// Apply the PWM duty cycle
	if !m.DisablePID {
		m.pwmDuty = m.pid.State.ControlSignal
	} else {
		m.pwmDuty = m.speedTable[m.currentSpeed]
	}
	m.ApplyPWM(m.pwmDuty)

	// Update the last control time
	m.lastControlTime = now
}

// SpeedMode returns the current DCC speed mode (14, 28, or 128 steps)
func (m *Motor) SpeedMode() SpeedMode {
	return m.speedMode
}

// SetSpeedMode sets the DCC speed mode (14, 28, or 128 steps)
func (m *Motor) SetSpeedMode(mode SpeedMode) {
	if m.speedMode == mode {
		return
	}
	m.speedMode = mode
	m.calculateAccelDecelRates()
	m.updateSpeedTable()
	m.updateBackEMFTiming()
}

func (m *Motor) SetSpeed(speed uint8, reverse bool) {
	// Update direction
	m.changeDirection = m.reverse != reverse

	// Handle emergency stop (speed value 1)
	if speed == 1 {
		m.emergencyStop()
		return
	}

	if m.targetSpeed == speed && !m.changeDirection {
		return
	}

	// If we're changing direction and currently moving, slow down and stop first
	if !m.changeDirection {
		m.setTargetSpeed(speed)
		m.updateBackEMFTiming()
	} else if m.currentSpeed > 0 && m.targetSpeed > 0 {
		m.setTargetSpeed(0)
		m.speedAfterStop = speed
	}
}

// stopMotor stops the motor
func (m *Motor) stopMotor() {
	m.ApplyPWM(0.0)
}

// emergencyStop immediately stops the motor
func (m *Motor) emergencyStop() {
	// For emergency stop, we want to stop immediately
	m.stopMotor()
	m.currentSpeed = 0
	m.rawSpeed = 0.0
	m.setTargetSpeed(0)
	m.pid.Reset()
	m.updatePIDConfig()
}

// updateSpeedStep adjusts the current speed step up or down based on max accel/decel rates
func (m *Motor) updateSpeedStep(elapsed time.Duration) {
	// If we're stopped and set to change direction, set the new speed
	if m.changeDirection && m.targetSpeed == 0 && m.currentSpeed == 0 {
		m.reverse = !m.reverse
		m.setTargetSpeed(m.speedAfterStop)
		m.speedAfterStop = 0
		m.changeDirection = false
		m.pid.Reset()
		m.setupADC(m.Direction())
	}

	// TODO: Handle going from 0 to non-zero speed after startup from dirty rail
	if m.currentSpeed > m.targetSpeed && m.rawSpeed-m.targetRaw > 0.5 {
		fmt.Printf("slow down - current: %d target: %d duty: %0.2f\r\n", m.currentSpeed, m.targetSpeed, m.speedTable[m.targetSpeed]) // TODO: Cleanup
		if m.decelRate > 0 {
			// decelRate: seconds per step * elapsed seconds = steps decreased
			m.rawSpeed -= m.decelRate * float32(elapsed.Seconds())
		} else {
			// If acceleration rate is zero, change to targetSpeed immediately
			m.rawSpeed = m.targetRaw
		}
	} else if m.currentSpeed < m.targetSpeed && m.targetRaw-m.rawSpeed > 0.5 {
		fmt.Printf("speed up - current: %d target: %d duty: %0.2f\r\n", m.currentSpeed, m.targetSpeed, m.speedTable[m.targetSpeed]) // TODO: Cleanup
		if m.accelRate > 0 {
			// accelRate: seconds per step * elapsed seconds = steps increased
			m.rawSpeed += m.accelRate * float32(elapsed.Seconds())
		} else {
			// If acceleration rate is zero, change to targetSpeed immediately
			m.rawSpeed = m.targetRaw
		}
	}
	// Round rawSpeed to nearest integer for speed table index
	m.currentSpeed = uint8(m.rawSpeed + 0.5)
}

// Make sure we always set the target speed and raw speed together
func (m *Motor) setTargetSpeed(speed uint8) {
	m.targetSpeed = speed
	m.targetRaw = float32(speed)
}

func (m *Motor) Direction() Direction {
	if m.reverse == m.ndotReverse {
		return Forward
	}
	return Reverse
}
