//go:build rp

package motor

import (
	"machine"
	"runtime"
	"time"

	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/hal"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/iir"
	"github.com/mikesmitty/tinypid"
)

type PWM interface {
	SetDuty(float32)
	SetFreq(uint64)
}

type Motor struct {
	// H-bridge PWM outputs
	pwmA      PWM
	pwmB      PWM
	pwmShared bool

	// Back-EMF measurement state for estimating speed
	emfA        *hal.ADC
	emfB        *hal.ADC
	adcRef      *hal.ADC
	emfDuration time.Duration
	emfInterval time.Duration
	emfMax      float32
	emfReading  float32
	emfTarget   float32
	emfTrigger  chan struct{}
	// TODO: Are these necessary?
	adcOffsetA float32
	adcOffsetB float32

	// IIR filters to smooth back-EMF readings
	iirAlpha float32
	iirA     *iir.IIRFilter
	iirB     *iir.IIRFilter
	iirRef   *iir.IIRFilter

	// Cached CV values
	cv map[uint16]uint8

	// Motor control state
	ndotReverse     bool
	reverse         bool
	currentSpeed    uint8   // Current speed step/speed table index
	rawSpeed        float32 // currentSpeed without rounding for accel/decel rates
	changeDirection bool    // Flag to indicate direction change commanded
	speedAfterStop  uint8   // Next speed step after direction change
	targetSpeed     uint8   // Commanded speed step
	speedMode       SpeedMode
	accelRate       float32 // Max steps per second accel
	decelRate       float32 // Max steps per second decel
	fwdTrim         float32
	revTrim         float32

	// For PID control
	pid       tinypid.PIController
	kpCutover uint8 // Speed step at which to switch from low to high PID gains
	kpLow     float32
	kpHigh    float32

	// Speed table for the 128 speed steps (including stop and e-stop)
	speedTable    [128]float32
	useSpeedTable bool

	// For sampling and control
	lastControlTime time.Time
}

func NewMotor(conf cv.Handler, hw *hal.HAL, pinA, pinB, emfA, emfB, adcRef machine.Pin) *Motor {
	m := &Motor{
		cv:              make(map[uint16]uint8),
		iirAlpha:        0.7,
		lastControlTime: time.Now(),
	}

	// Set up motor driver pins for PWM
	err := m.initPWM(hw, pinA, pinB, hal.MaxMotorPWMFreq, 0.0)
	if err != nil {
		panic(err.Error())
	}

	// Set up back EMF pins for ADC
	m.initBackEMF(emfA, emfB, adcRef)

	// Initialize PID controller
	m.pid = tinypid.PIController{
		Config: tinypid.PIControllerConfig{
			MaxIntegralError: 100,
			MinIntegralError: -100,
			MaxOutput:        1.0,
			MinOutput:        0,
		},
	}

	// Initialize speed table
	m.updateSpeedTable()

	// Calculate acceleration/deceleration rates
	m.calculateAccelDecelRates()

	return m
}

// Run the motor controller loop
func (m *Motor) Run() {
	for {
		select {
		case <-time.After(100 * time.Millisecond):
			m.runMotorControl()
		case <-m.emfTrigger:
			m.emfReading = m.measureBackEMF()
		default:
			runtime.Gosched()
		}
	}
}

// SpeedMode returns the current DCC speed mode (14, 28, or 128 steps)
func (m *Motor) SpeedMode() SpeedMode {
	return m.speedMode
}

// SetSpeedMode sets the DCC speed mode (14, 28, or 128 steps)
func (m *Motor) SetSpeedMode(mode SpeedMode) {
	m.speedMode = mode
	m.calculateAccelDecelRates()
	m.updateSpeedTable()
}

// SetSpeed sets the motor speed based on DCC speed steps
func (m *Motor) SetSpeed(speed uint8, reverse bool) {
	// Update direction
	m.changeDirection = m.reverse != reverse
	m.reverse = reverse

	// Handle emergency stop (speed value 1)
	if speed == 1 {
		m.emergencyStop()
		return
	}

	// If we're changing direction and currently moving, slow down and stop first
	if m.changeDirection && m.currentSpeed > 0 {
		m.targetSpeed = 0
		m.speedAfterStop = speed
	} else {
		m.targetSpeed = speed
	}
}

// stopMotor stops the motor
func (m *Motor) stopMotor() {
	m.applyPWM(0.0)
}

// emergencyStop immediately stops the motor
func (m *Motor) emergencyStop() {
	// For emergency stop, we want to stop immediately
	m.stopMotor()
	m.currentSpeed = 0
	m.rawSpeed = 0.0
	m.targetSpeed = 0
	m.pid.Reset()
	m.updatePIDConfig()
}

// runMotorControl is the main control loop for the motor
func (m *Motor) runMotorControl() {
	now := time.Now()
	elapsed := now.Sub(m.lastControlTime)

	// Update speed step
	prevSpeed := m.currentSpeed
	m.updateSpeedStep(elapsed)
	if prevSpeed != m.currentSpeed {
		// If the speed step changed, update the PID controller config and EMF target
		m.updatePIDConfig()
		m.emfTarget = m.speedTable[m.currentSpeed] * m.emfMax
	}

	// Run the PID loop
	m.pid.Update(tinypid.PIControllerInput{
		ReferenceSignal: m.emfTarget,
		ActualSignal:    m.emfReading,
	})
}

// updateSpeedStep steps the motor speed up or down based on max accel/decel rates
func (m *Motor) updateSpeedStep(elapsed time.Duration) {
	// If we're stopped and set to change direction, set the new speed
	if m.changeDirection && m.targetSpeed == 0 && m.currentSpeed == 0 {
		m.reverse = !m.reverse
		m.targetSpeed = m.speedAfterStop
		m.speedAfterStop = 0
		m.changeDirection = false
		m.pid.Reset()
	}

	if m.currentSpeed > m.targetSpeed && m.rawSpeed-float32(m.targetSpeed) > 0.5 {
		// decelRate: seconds per step * elapsed seconds = steps decreased
		m.rawSpeed -= m.decelRate * float32(elapsed.Seconds())
	} else if m.currentSpeed < m.targetSpeed && float32(m.targetSpeed)-m.rawSpeed > 0.5 {
		// accelRate: seconds per step * elapsed seconds = steps increased
		m.rawSpeed += m.accelRate * float32(elapsed.Seconds())
	}
	// Round rawSpeed to nearest integer for speed table index
	m.currentSpeed = uint8(m.rawSpeed + 0.5)
}
