package motor

import (
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
		{"ConstantSpeed", 128, 128, 128, SpeedMode28},
		{"ExtremeLowToHigh", 0, 255, 0, SpeedMode28},
		{"AllMaxValues", 255, 255, 255, SpeedMode28},
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

			// Ensure no speed step decreases compared to the previous value
			for i := 3; i < int(tt.speedMode)+2; i++ {
				if motor.speedTable[i] < motor.speedTable[i-1] {
					t.Errorf("Speed step %d decreased: %f < %f", i, motor.speedTable[i], motor.speedTable[i-1])
				}
			}

			// Additional checks for specific cases
			switch tt.name {
			case "ConstantSpeed":
				for i := 2; i < int(tt.speedMode)+2; i++ {
					if motor.speedTable[i] != motor.speedTable[2] {
						t.Errorf("Expected constant speed table, but got %f at step %d", motor.speedTable[i], i)
					}
				}
			case "ExtremeLowToHigh":
				if motor.speedTable[1] != 0 || motor.speedTable[tt.speedMode+1] != 1.0 {
					t.Errorf("Expected speed table to start at 0 and end at 1.0, but got %f to %f", motor.speedTable[2], motor.speedTable[tt.speedMode+1])
				}
			case "AllMaxValues":
				for i := 2; i < int(tt.speedMode)+2; i++ {
					if motor.speedTable[i] != 1.0 {
						t.Errorf("Expected all speed steps to be 1.0, but got %f at step %d", motor.speedTable[i], i)
					}
				}
			}
		})
	}
}
