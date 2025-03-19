package dcc

import (
	"testing"

	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
)

func TestSetCVCommand(t *testing.T) {
	tests := []struct {
		name   string
		buf    []byte
		mockCV cv.Handler
		expect bool
	}{
		{
			name:   "Verify byte success",
			buf:    []byte{0b0100, 10, 0x42}, // op=01, cv=10, data=0x42
			mockCV: cv.NewMockHandler(true, testCVConfirm([]uint16{10}, []uint8{0x42})),
			expect: true,
		},
		{
			name:   "Verify CV failure",
			buf:    []byte{0b0100, 10, 0x42}, // op=01, cv=10, data=0x42
			mockCV: cv.NewMockHandler(true, testCVConfirm([]uint16{10}, []uint8{0x43})),
			expect: false,
		},
		{
			name:   "Write byte success",
			buf:    []byte{0b1100, 10, 0x42}, // op=11, cv=10, data=0x42
			mockCV: cv.NewMockHandler(true, make(map[uint16]uint8)),
			expect: true,
		},
		{
			name:   "Write byte failure",
			buf:    []byte{0b1100, 10, 0x42}, // op=11, cv=10, data=0x42
			mockCV: cv.NewMockHandler(false, make(map[uint16]uint8)),
			expect: false,
		},
		{
			name:   "Verify bit success",
			buf:    []byte{0b1000, 10, 0b11101001}, // op=10, cv=10, verify bit 1 at position 1
			mockCV: cv.NewMockHandler(true, testCVConfirm([]uint16{10}, []uint8{2})),
			expect: true,
		},
		{
			name:   "Verify bit failure",
			buf:    []byte{0b1000, 10, 0b11101001}, // op=10, cv=10, verify bit 1 at position 1
			mockCV: cv.NewMockHandler(true, testCVConfirm([]uint16{10}, []uint8{0})),
			expect: false,
		},
		{
			name:   "Write bit success",
			buf:    []byte{0b1000, 10, 0b11111001}, // op=10, cv=10, write bit 1 at position 1
			mockCV: cv.NewMockHandler(true, testCVConfirm([]uint16{10}, []uint8{0})),
			expect: true,
		},
		{
			name:   "Write bit invalid CV",
			buf:    []byte{0b1000, 10, 0b11111001}, // op=10, cv=10, write bit 1 at position 1
			mockCV: cv.NewMockHandler(false, testCVConfirm([]uint16{10}, []uint8{0})),
			expect: false,
		},
		{
			name:   "Invalid bit manipulation command",
			buf:    []byte{0b1000, 10, 0x01}, // op=10, cv=10, invalid bit command (missing 111 prefix)
			mockCV: cv.NewMockHandler(true, testCVConfirm([]uint16{10}, []uint8{0})),
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{cv: tt.mockCV}
			got := m.setCVCommand(0, tt.buf)
			if got != tt.expect {
				t.Errorf("setCVCommand() = %v, want %v", got, tt.expect)
			}
		})
	}
}
