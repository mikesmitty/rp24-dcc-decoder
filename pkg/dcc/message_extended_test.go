package dcc

import (
	"testing"

	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/motor"
)

func testCVConfirm(cvs []uint16, values []uint8) map[uint16]uint8 {
	cvConfirm := make(map[uint16]uint8)
	for i := range cvs {
		cvConfirm[cvs[i]] = values[i]
	}
	return cvConfirm
}

func TestExtendedPacket(t *testing.T) {
	tests := []struct {
		name   string
		msg    *Message
		expect bool
	}{
		{
			name: "Decoder Consist Control Instruction",
			msg: &Message{
				addr:    BroadcastAddress,
				buf:     []byte{0x00, 0x00, 0x00},
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			expect: true,
		},
		{
			name: "Advanced Operation Instruction",
			msg: &Message{
				buf:     []byte{0x03, 0x3F, 0x00, 0xFF},
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			expect: true,
		},
		{
			name: "Speed and Direction Instruction Forward",
			msg: &Message{
				buf:     []byte{0x03, 0x60, 0xFF},
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			expect: true,
		},
		{
			name: "Function Group One Instruction",
			msg: &Message{
				buf:     []byte{0x03, 0x80, 0x00, 0xFF},
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			expect: true,
		},
		{
			name: "Function Group Two Instruction F5-F8",
			msg: &Message{
				buf:     []byte{0x03, 0xA0, 0xFF},
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			expect: true,
		},
		{
			name: "Function Group Two Instruction F9-F12",
			msg: &Message{
				buf:     []byte{0x03, 0xB0, 0xFF},
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			expect: true,
		},
		{
			name: "Feature Expansion",
			msg: &Message{
				buf:     []byte{0x03, 0xC0, 0x00, 0xFF},
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			expect: false,
		},
		{
			// Long form - Set CV
			name: "Config Variable Access Instruction",
			msg: &Message{
				buf:     []byte{0x03, 0xE0, 0x00, 0xFF},
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			expect: false,
		},
		{
			name: "Config Variable Access DirectAddress Edit Check",
			msg: &Message{
				addr:    DirectAddress,
				buf:     []byte{0x03, 0xEC, 0x03, 0x01, 0xFF}, // Write 0x01 to CV3
				cv:      cv.NewMockHandler(true, make(map[uint16]uint8)),
				decoder: &Decoder{address: []byte{3}, consistAddress: []byte{1}, motor: &motor.Motor{}},
			},
			expect: true,
		},
		{
			name: "Config Variable Access ConsistAddress Edit Check",
			msg: &Message{
				addr:    ConsistAddress,
				buf:     []byte{0x01, 0xEC, 0x03, 0x01, 0xFF}, // Write 0x01 to CV3
				cv:      cv.NewMockHandler(true, make(map[uint16]uint8)),
				decoder: &Decoder{address: []byte{3}, consistAddress: []byte{1}, motor: &motor.Motor{}},
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.extendedPacket()
			if result != tt.expect {
				t.Errorf("expected %v, got %v", tt.expect, result)
			}
		})
	}
}

func TestDecoderConsistControlInstruction(t *testing.T) {
	tests := []struct {
		name   string
		msg    *Message
		input  []byte
		expect bool
	}{
		{
			name: "Decoder Reset Packet",
			msg: &Message{
				addr:    BroadcastAddress,
				decoder: &Decoder{address: []byte{3}},
			},
			input:  []byte{0x00},
			expect: true,
		},
		{
			name: "Decoder Hard Reset Packet",
			msg: &Message{
				cv:      cv.NewMockHandler(true, make(map[uint16]uint8)),
				decoder: &Decoder{address: []byte{3}},
			},
			input:  []byte{0x01},
			expect: true,
		},
		{
			name: "Set Advanced Addressing Mode",
			msg: &Message{
				cv:      cv.NewMockHandler(true, make(map[uint16]uint8)),
				decoder: &Decoder{address: []byte{3}},
			},
			input:  []byte{0x0A},
			expect: true,
		},
		{
			name: "Set Consist Address",
			msg: &Message{
				cv:      cv.NewMockHandler(true, make(map[uint16]uint8)),
				decoder: &Decoder{address: []byte{3}, consistAddress: []byte{0}},
			},
			input:  []byte{0x12, 0x01},
			expect: true,
		},
		{
			name: "Set Consist Address and Reverse Direction",
			msg: &Message{
				cv:      cv.NewMockHandler(true, make(map[uint16]uint8)),
				decoder: &Decoder{address: []byte{3}, consistAddress: []byte{0}},
			},
			input:  []byte{0x13, 0x01},
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.decoderConsistControlInstruction(tt.input)
			if result != tt.expect {
				t.Errorf("expected %v, got %v", tt.expect, result)
			}
		})
	}
}

func TestAdvancedOperationInstruction(t *testing.T) {
	tests := []struct {
		name   string
		msg    *Message
		input  []byte
		expect bool
	}{
		{
			name: "128-step speed control",
			msg: &Message{
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			input:  []byte{0x3F, 0x00},
			expect: true,
		},
		{
			name: "Invalid command",
			msg: &Message{
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			input:  []byte{0x00},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.advancedOperationInstruction(tt.input)
			if result != tt.expect {
				t.Errorf("expected %v, got %v", tt.expect, result)
			}
		})
	}
}

func TestFunctionGroupOneInstruction(t *testing.T) {
	msg := &Message{
		decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
	}
	result := msg.functionGroupOneInstruction(0x10)
	if !result {
		t.Errorf("expected true, got %v", result)
	}
}

func TestFunctionGroupTwoInstruction(t *testing.T) {
	msg := &Message{
		decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
	}
	result := msg.functionGroupTwoInstruction(0x10)
	if !result {
		t.Errorf("expected true, got %v", result)
	}
}

func TestFeatureExpansion(t *testing.T) {
	tests := []struct {
		name   string
		msg    *Message
		input  []byte
		expect bool
	}{
		{
			name: "Functions F13-F20",
			msg: &Message{
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			input:  []byte{0xDE, 0x01},
			expect: true,
		},
		{
			name: "Functions F21-F28",
			msg: &Message{
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			input:  []byte{0xDF, 0x01},
			expect: true,
		},
		{
			name: "Functions F29-F36",
			msg: &Message{
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			input:  []byte{0xD8, 0x01},
			expect: true,
		},
		{
			name: "Functions F37-F44",
			msg: &Message{
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			input:  []byte{0xD9, 0x01},
			expect: true,
		},
		{
			name: "Functions F45-F52",
			msg: &Message{
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			input:  []byte{0xDA, 0x01},
			expect: true,
		},
		{
			name: "Functions F53-F61",
			msg: &Message{
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			input:  []byte{0xDB, 0x01},
			expect: true,
		},
		{
			name: "Functions F62-F68",
			msg: &Message{
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			input:  []byte{0xDC, 0x01},
			expect: true,
		},
		{
			name: "Invalid command",
			msg: &Message{
				decoder: &Decoder{address: []byte{3}, motor: &motor.Motor{}},
			},
			input:  []byte{0x00},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.featureExpansion(tt.input)
			if result != tt.expect {
				t.Errorf("expected %v, got %v", tt.expect, result)
			}
		})
	}
}

func TestConfigVariableAccessInstruction(t *testing.T) {
	tests := []struct {
		name   string
		msg    *Message
		input  []byte
		expect bool
	}{
		// Long form commands
		{
			name:   "Invalid XPOM command",
			msg:    &Message{cv: cv.NewMockHandler(true, make(map[uint16]uint8))},
			input:  []byte{0xE0, 0x00, 0x00, 0x01},
			expect: false,
		},
		// Short form commands
		{
			name:   "CV23 Acceleration rate adjustment",
			msg:    &Message{cv: cv.NewMockHandler(true, make(map[uint16]uint8))},
			input:  []byte{0xF2, 0x01},
			expect: true,
		},
		{
			name:   "CV24 Deceleration rate adjustment",
			msg:    &Message{cv: cv.NewMockHandler(true, make(map[uint16]uint8))},
			input:  []byte{0xF3, 0x01},
			expect: true,
		},
		{
			name:   "Extended address programming invalid command length",
			msg:    &Message{cv: cv.NewMockHandler(true, make(map[uint16]uint8))},
			input:  []byte{0xF4, 0x01},
			expect: false,
		},
		{
			name:   "Extended address programming first message",
			msg:    &Message{cv: cv.NewMockHandler(true, make(map[uint16]uint8)), cvConfirm: make(map[uint16]byte)},
			input:  []byte{0xF4, 0x00, 0x00},
			expect: true,
		},
		{
			name:   "Extended address programming second message",
			msg:    &Message{cv: cv.NewMockHandler(true, make(map[uint16]uint8)), cvConfirm: testCVConfirm([]uint16{17, 18}, []uint8{0xC0, 0x05})},
			input:  []byte{0xF4, 0xC0, 0x05},
			expect: true,
		},
		{
			name:   "Indexed CVs invalid command length",
			msg:    &Message{cv: cv.NewMockHandler(true, make(map[uint16]uint8))},
			input:  []byte{0xF5, 0x01},
			expect: false,
		},
		{
			name:   "Indexed CVs reserved CV31",
			msg:    &Message{cv: cv.NewMockHandler(true, make(map[uint16]uint8))},
			input:  []byte{0xF5, 0x01, 0x00},
			expect: false,
		},
		{
			name:   "Indexed CVs first message",
			msg:    &Message{cv: cv.NewMockHandler(true, make(map[uint16]uint8)), cvConfirm: make(map[uint16]byte)},
			input:  []byte{0xF5, 0x10, 0x00},
			expect: true,
		},
		{
			name:   "Indexed CVs second message",
			msg:    &Message{cv: cv.NewMockHandler(true, make(map[uint16]uint8)), cvConfirm: testCVConfirm([]uint16{31, 32}, []uint8{0x10, 0x00})},
			input:  []byte{0xF5, 0x10, 0x00},
			expect: true,
		},
		{
			name:   "Consist extended address invalid command length",
			msg:    &Message{cv: cv.NewMockHandler(true, make(map[uint16]uint8))},
			input:  []byte{0xF6, 0x01},
			expect: false,
		},
		{
			name:   "Consist extended address first message",
			msg:    &Message{cv: cv.NewMockHandler(true, make(map[uint16]uint8)), cvConfirm: make(map[uint16]byte)},
			input:  []byte{0xF6, 0x00, 0x01},
			expect: true,
		},
		{
			name:   "Consist extended address second message",
			msg:    &Message{cv: cv.NewMockHandler(true, make(map[uint16]uint8)), cvConfirm: testCVConfirm([]uint16{19, 20}, []uint8{0x00, 0x01})},
			input:  []byte{0xF6, 0x00, 0x01},
			expect: true,
		},
		{
			name:   "Invalid command",
			msg:    &Message{},
			input:  []byte{0x00},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.configVariableAccessInstruction(tt.input)
			if result != tt.expect {
				t.Errorf("expected %v, got %v", tt.expect, result)
			}
		})
	}
}
