package dcc

import (
	"testing"

	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/motor"
)

func FuzzMessage(f *testing.F) {
	testcases := [][]byte{
		{0b00000000, 0b00000000, 0b00000000},
		{0b01111111, 0b10000000},
		{0x03, 0x3F, 0x00},
		{0x03, 0x60},
		{0x03, 0x80, 0x00},
	}
	for _, tc := range testcases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}
	f.Fuzz(func(t *testing.T, msg []byte) {
		if len(msg) > 8 {
			// Limit the length of message fuzzing, most all messages are 6 bytes or less
			return
		}

		m := NewMessage(cv.NewMockHandler(true, make(map[uint16]uint8)), &Decoder{
			address:        []byte{3},
			consistAddress: []byte{1},
			motor:          &motor.Motor{},
		})
		m.buf = msg

		// Generate a valid checksum
		var xor byte
		for _, b := range msg {
			xor ^= b
		}
		m.AddByte(xor)

		m.Process()
	})
}
