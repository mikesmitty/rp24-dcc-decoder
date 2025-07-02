package motor

import (
	"fmt"
	"testing"
)

func TestGenerate3PointSpeedTable(t *testing.T) {
	tests := []struct {
		name      string
		cv2       uint8 // Vstart
		cv5       uint8 // Vmax
		cv6       uint8 // Vmid
		speedMode SpeedMode
	}{
		{"LinearIncrease", 64, 192, 128, SpeedMode28},
		{"DefaultMidpoint", 64, 192, 0, SpeedMode28},
		{"MaxZero", 64, 0, 128, SpeedMode28},
		{"ExtremeLowToHigh", 0, 255, 0, SpeedMode28},
		{"SpeedMode14", 64, 192, 128, SpeedMode14},
		{"SpeedMode128", 64, 192, 128, SpeedMode128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			motor := &Motor{
				cv:        make(map[uint16]uint8, 256),
				speedMode: SpeedMode(tt.speedMode),
			}
			motor.cv[2] = tt.cv2
			motor.cv[5] = tt.cv5
			motor.cv[6] = tt.cv6

			motor.generate3PointSpeedTable()

			printThrottleGraph(120, motor)

			// Ensure no speed step decreases or plateaus compared to the previous value
			for i := 3; i < int(tt.speedMode)+2; i++ {
				if motor.speedTable[i] <= motor.speedTable[i-1] {
					t.Errorf("Speed step %d did not increase: %f <= %f", i, motor.speedTable[i], motor.speedTable[i-1])
				}
			}

			// Additional checks for specific cases
			switch tt.name {
			case "ExtremeLowToHigh":
				if motor.speedTable[1] != 0 || motor.speedTable[tt.speedMode+1] != 1.0 {
					t.Errorf("Expected speed table to start at 0 and end at 1.0, but got %f to %f", motor.speedTable[2], motor.speedTable[tt.speedMode+1])
				}
			}
		})
	}
}

func printThrottleGraph(width int, m *Motor) {
	for i := range int(m.speedMode) + 2 {
		steps := int(float32(width-6) * m.speedTable[i])
		fmt.Printf("%3d: ", i)
		for range steps {
			print("#")
		}
		fmt.Printf(" %0.3f\n", m.speedTable[i])
	}
}
