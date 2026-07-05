package motor

import (
	"math"
	"testing"
	"time"

	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/hal"
)

func TestBEMFSpeedControlSimulation(t *testing.T) {
	// Reset the global hooks at the end of the test
	defer func() {
		hal.ADCReadHook = nil
		hal.PWMInitHook = nil
		hal.PWMSetDutyHook = nil
	}()

	// 1. Create a CV store with realistic BEMF settings
	cvStore := map[uint16]uint8{
		2:   10,          // Vstart
		5:   255,         // Vmax
		6:   128,         // Vmid
		9:   40,          // PWM Frequency 40kHz
		29:  0b00000010,  // 28-speed mode
		49:  1,           // DisablePID = false (meaning PID is ENABLED!)
		50:  20,          // emfSettle: 20 * 5us = 100us
		51:  14,          // kpCutover speed step 14
		52:  20,          // kpLow: 2.0
		54:  30,          // kpHigh: 3.0
		55:  10,          // Ki: 1.0
		53:  90,          // emfMax: 9.0V max-speed BEMF (N scale motor on ~12V track)
		116: 100,         // emfInterval: 100 * 100us = 10ms
		117: 100,
		118: 20,          // emfDuration: 20 * 100us = 2ms
		119: 20,
	}
	mockCV := cv.NewMockHandler(true, cvStore)

	// 2. Initialize a mock HAL and the Motor
	hw := hal.NewHAL()
	pinA := shared.MockPin(1)
	pinB := shared.MockPin(2)
	emfA := shared.MockPin(3)
	emfB := shared.MockPin(4)

	// Simulated physical motor variables
	var simSpeed float32 = 0.0
	var dutyA, dutyB float32
	// Motor voltage constant: simSpeed 1.0 produces the full 9V max-speed BEMF,
	// converted to ADC counts through the board's 6.8k/1k sense divider (~2546 counts/V)
	const maxBEMFVolts float32 = 9.0
	const Kv float32 = maxBEMFVolts * bemfCountsPerVolt
	const dt float32 = 0.1        // Simulation time step (100ms matches pwmInterval/control loop)
	const mass float32 = 0.5      // Virtual rotor mass / inertia
	const drag float32 = 0.2      // Simulated rotational friction / drag
	var virtualLoad float32 = 0.0 // External torque/resistance

	// Track which SimplePWM instance drives which H-bridge side so the SetDuty
	// hook can tell them apart (ApplyPWM writes the driven side and zeroes the other)
	pwmPin := make(map[*hal.SimplePWM]shared.Pin)
	hal.PWMInitHook = func(pin shared.Pin, freq uint64, duty float32) (*hal.SimplePWM, error) {
		p := &hal.SimplePWM{}
		pwmPin[p] = pin
		return p, nil
	}
	hal.PWMSetDutyHook = func(p *hal.SimplePWM, duty float32) {
		switch pwmPin[p] {
		case pinA:
			dutyA = duty
		case pinB:
			dutyB = duty
		}
	}

	m := NewMotor(mockCV, hw, pinA, pinB, emfA, emfB)
	m.SetSpeedMode(SpeedMode28)

	// Effective forward drive duty applied to the simulated motor
	currentDuty := func() float32 {
		return dutyA - dutyB
	}

	// Hook ADC Read to return back EMF proportional to virtual speed
	hal.ADCReadHook = func(pin shared.Pin) uint16 {
		// Back EMF is proportional to simulated motor speed
		bemfRaw := simSpeed * Kv
		if bemfRaw < 0 {
			bemfRaw = 0
		}
		if bemfRaw > 65535 {
			bemfRaw = 65535
		}
		return uint16(bemfRaw)
	}

	// 3. Request a target speed step
	m.SetSpeed(14, false) // Command mid-speed (speed step 14)

	// 4. Run the simulation loop for 50 steps
	t.Logf("Ramping up to speed step 14...")
	var speedHistory []float32
	var dutyHistory []float32

	for range 50 {
		// Calculate Back EMF measurement (measureBackEMF updates m.emfValue)
		m.emfValue = m.measureBackEMF()

		// Run the motor control loop which updates m.pwmDuty and applies PWM
		m.runMotorControl()

		// Simulate the physical motor physics:
		// torque = drive duty - drag * simSpeed - virtualLoad
		// acceleration = torque / mass
		// simSpeed = simSpeed + acceleration * dt
		torque := currentDuty() - drag*simSpeed - virtualLoad
		simSpeed += (torque / mass) * dt
		if simSpeed < 0 {
			simSpeed = 0
		}

		speedHistory = append(speedHistory, simSpeed)
		dutyHistory = append(dutyHistory, currentDuty())
	}

	finalSpeed := speedHistory[len(speedHistory)-1]
	finalDuty := dutyHistory[len(dutyHistory)-1]
	t.Logf("After 50 steps, virtual speed: %f, PWM duty cycle: %f", finalSpeed, finalDuty)

	// Verify that speed control is active and has stabilized the motor to a non-zero speed
	if finalSpeed < 0.1 {
		t.Errorf("Motor did not start or speed is too low: %f", finalSpeed)
	}
	if finalDuty < 0.05 {
		t.Errorf("Duty cycle is too low, speed control is not driving the motor: %f", finalDuty)
	}

	// 5. Apply load and verify compensation!
	t.Logf("Applying external load/friction...")
	virtualLoad = 0.15 // Apply friction/drag load

	for range 30 {
		m.emfValue = m.measureBackEMF()
		m.runMotorControl()

		torque := currentDuty() - drag*simSpeed - virtualLoad
		simSpeed += (torque / mass) * dt
		if simSpeed < 0 {
			simSpeed = 0
		}

		speedHistory = append(speedHistory, simSpeed)
		dutyHistory = append(dutyHistory, currentDuty())
	}

	postLoadSpeed := speedHistory[len(speedHistory)-1]
	postLoadDuty := dutyHistory[len(dutyHistory)-1]
	t.Logf("After applying load, virtual speed: %f (was %f), PWM duty: %f (was %f)",
		postLoadSpeed, finalSpeed, postLoadDuty, finalDuty)

	// Speed should recover/stay stable, and duty cycle MUST increase to compensate for load!
	if postLoadDuty <= finalDuty {
		t.Errorf("PID controller failed to compensate! Duty cycle did not increase under load (was %f, now %f)", finalDuty, postLoadDuty)
	}
	if math.Abs(float64(postLoadSpeed-finalSpeed)) > 0.1 {
		t.Logf("Warning: large speed deviation, but PID compensated duty cycle from %f to %f", finalDuty, postLoadDuty)
	}
}

// Verify the driver autosleep wake blip: after a back-EMF cutout, the driven
// output must be held solid-high before normal PWM resumes (needed for
// autosleep H-bridges like the DRV8220), and skipped when not configured
func TestMeasureBackEMFDriverWake(t *testing.T) {
	defer func() {
		hal.ADCReadHook = nil
		hal.PWMInitHook = nil
		hal.PWMSetDutyHook = nil
	}()

	newTestMotor := func() (*Motor, *[]float32) {
		cvStore := map[uint16]uint8{29: 0b00000010}
		mockCV := cv.NewMockHandler(true, cvStore)
		hw := hal.NewHAL()
		pinA := shared.MockPin(1)
		pinB := shared.MockPin(2)

		pwmPin := make(map[*hal.SimplePWM]shared.Pin)
		var dutiesA []float32
		hal.PWMInitHook = func(pin shared.Pin, freq uint64, duty float32) (*hal.SimplePWM, error) {
			p := &hal.SimplePWM{}
			pwmPin[p] = pin
			return p, nil
		}
		hal.PWMSetDutyHook = func(p *hal.SimplePWM, duty float32) {
			if pwmPin[p] == pinA {
				dutiesA = append(dutiesA, duty)
			}
		}

		m := NewMotor(mockCV, hw, pinA, pinB, shared.MockPin(3), shared.MockPin(4))
		return m, &dutiesA
	}

	t.Run("wake blip enabled", func(t *testing.T) {
		m, dutiesA := newTestMotor()
		m.DriverWakeTime = 100 * time.Microsecond
		m.pwmDuty = 0.4

		m.measureBackEMF()

		d := *dutiesA
		if len(d) < 3 {
			t.Fatalf("expected at least 3 duty writes (cutout, wake, restore), got %v", d)
		}
		last3 := d[len(d)-3:]
		if last3[0] != 0.0 || last3[1] != 1.0 || last3[2] != 0.4 {
			t.Errorf("expected duty sequence [0 1 0.4] (cutout, wake blip, restore), got %v", last3)
		}
	})

	t.Run("wake blip disabled", func(t *testing.T) {
		m, dutiesA := newTestMotor()
		m.pwmDuty = 0.4

		m.measureBackEMF()

		for _, duty := range *dutiesA {
			if duty == 1.0 {
				t.Errorf("unexpected full-duty wake blip with DriverWakeTime unset: %v", *dutiesA)
			}
		}
	})
}
