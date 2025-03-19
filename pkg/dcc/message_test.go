package dcc

import (
	"reflect"
	"testing"

	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/motor"
)

func TestNewMessage(t *testing.T) {
	cvHandler := &cv.CVHandler{}
	decoder := &Decoder{}

	msg := NewMessage(cvHandler, decoder)

	if msg == nil {
		t.Fatal("NewMessage returned nil")
	}

	if msg.cv != cvHandler {
		t.Errorf("Expected cv handler to be set")
	}

	if msg.decoder != decoder {
		t.Errorf("Expected decoder to be set")
	}

	if cap(msg.buf) != 11 {
		t.Errorf("Expected buffer capacity to be 11, got %d", cap(msg.buf))
	}

	if len(msg.buf) != 0 {
		t.Errorf("Expected buffer length to be 0, got %d", len(msg.buf))
	}
}

func TestAddByte(t *testing.T) {
	msg := NewMessage(nil, nil)

	msg.AddByte(0x01)
	if len(msg.buf) != 1 || msg.buf[0] != 0x01 {
		t.Errorf("AddByte failed to add byte properly")
	}

	msg.AddByte(0x02)
	if len(msg.buf) != 2 || msg.buf[1] != 0x02 {
		t.Errorf("AddByte failed to add second byte properly")
	}
}

func TestAddBytes(t *testing.T) {
	msg := NewMessage(nil, nil)

	bytes := []byte{0x01, 0x02, 0x03}
	msg.AddBytes(bytes)

	if len(msg.buf) != 3 {
		t.Errorf("AddBytes failed to add bytes properly, expected length 3, got %d", len(msg.buf))
	}

	if !reflect.DeepEqual(msg.buf, bytes) {
		t.Errorf("AddBytes failed to add correct bytes, expected %v, got %v", bytes, msg.buf)
	}
}

func TestBytes(t *testing.T) {
	msg := NewMessage(nil, nil)

	bytes := []byte{0x01, 0x02, 0x03}
	msg.AddBytes(bytes)

	result := msg.Bytes()

	if !reflect.DeepEqual(result, bytes) {
		t.Errorf("Bytes failed to return correct bytes, expected %v, got %v", bytes, result)
	}
}

func TestReset(t *testing.T) {
	msg := NewMessage(nil, nil)

	msg.AddBytes([]byte{0x01, 0x02, 0x03})
	msg.msgType = ServiceMsg

	msg.Reset()

	if len(msg.buf) != 0 {
		t.Errorf("Reset failed to clear buffer, length is %d", len(msg.buf))
	}

	if msg.msgType != UnknownMsg {
		t.Errorf("Reset failed to reset message type")
	}
}

func TestXOR(t *testing.T) {
	msg := NewMessage(nil, nil)

	// Test with valid XOR (XOR of all bytes is 0)
	msg.AddBytes([]byte{0x01, 0x02, 0x03})
	// XOR of 0x01 ^ 0x02 ^ 0x03 is 0x00
	// Add to the XOR result to make it invalid
	msg.AddByte(0x05)

	if msg.XOR() {
		t.Errorf("XOR should return false for invalid checksum")
	}

	// Test with correct XOR
	msg.Reset()
	msg.AddBytes([]byte{0x01, 0x02, 0x03})
	// XOR of 0x01 ^ 0x02 ^ 0x03 is 0x00
	// Add the XOR result to make the total XOR 0
	msg.AddByte(0x00 ^ 0x01 ^ 0x02 ^ 0x03)

	if !msg.XOR() {
		t.Errorf("XOR should return true for valid checksum")
	}
}

func TestIsEmpty(t *testing.T) {
	msg := NewMessage(nil, nil)

	if !msg.IsEmpty() {
		t.Errorf("IsEmpty should return true for new message")
	}

	msg.AddByte(0x01)

	if msg.IsEmpty() {
		t.Errorf("IsEmpty should return false after adding bytes")
	}
}

func TestIsFull(t *testing.T) {
	msg := NewMessage(nil, nil)

	if msg.IsFull() {
		t.Errorf("IsFull should return false for new message")
	}

	// Fill the buffer to capacity
	for i := 0; i < cap(msg.buf); i++ {
		msg.AddByte(byte(i))
	}

	if !msg.IsFull() {
		t.Errorf("IsFull should return true when buffer is at capacity")
	}
}

func TestLength(t *testing.T) {
	msg := NewMessage(nil, nil)

	if msg.Length() != 0 {
		t.Errorf("Length should return 0 for new message")
	}

	msg.AddBytes([]byte{0x01, 0x02, 0x03})

	if msg.Length() != 3 {
		t.Errorf("Length should return 3 after adding 3 bytes")
	}
}

func TestAddressMatch(t *testing.T) {
	msg := NewMessage(nil, nil)
	msg.AddBytes([]byte{0x01, 0x02, 0x03})

	tests := []struct {
		name    string
		address []byte
		want    bool
	}{
		{
			name:    "Exact match",
			address: []byte{0x01, 0x02, 0x03},
			want:    true,
		},
		{
			name:    "Prefix match",
			address: []byte{0x01, 0x02},
			want:    true,
		},
		{
			name:    "Single byte match",
			address: []byte{0x01},
			want:    true,
		},
		{
			name:    "No match",
			address: []byte{0x02, 0x03, 0x04},
			want:    false,
		},
		{
			name:    "Too long",
			address: []byte{0x01, 0x02, 0x03, 0x04},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := msg.addressMatch(tt.address); got != tt.want {
				t.Errorf("addressMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageType(t *testing.T) {
	tests := []struct {
		name         string
		buffer       []byte
		opMode       opMode
		svcModeReady bool
		expectedType MessageType
	}{
		{
			name:         "Service mode packet",
			buffer:       []byte{0x70}, // 0x70 = 112 decimal
			opMode:       ServiceMode,
			svcModeReady: false,
			expectedType: ServiceMsg,
		},
		{
			name:         "Service mode packet with svcModeReady",
			buffer:       []byte{0x7F}, // 0x7F = 127 decimal
			opMode:       OperationsMode,
			svcModeReady: true,
			expectedType: ServiceMsg,
		},
		{
			name:         "Extended message format (0-127)",
			buffer:       []byte{0x03},
			opMode:       OperationsMode,
			svcModeReady: false,
			expectedType: ExtendedMsg,
		},
		{
			name:         "Extended message format (192-231)",
			buffer:       []byte{0xC0}, // 0xC0 = 192 decimal
			opMode:       OperationsMode,
			svcModeReady: false,
			expectedType: ExtendedMsg,
		},
		{
			name:         "Advanced extended message format (253)",
			buffer:       []byte{0xFD}, // 0xFD = 253 decimal
			opMode:       OperationsMode,
			svcModeReady: false,
			expectedType: AdvancedExtendedMsg,
		},
		{
			name:         "Advanced extended message format (254)",
			buffer:       []byte{0xFE}, // 0xFE = 254 decimal
			opMode:       OperationsMode,
			svcModeReady: false,
			expectedType: AdvancedExtendedMsg,
		},
		{
			name:         "Unknown message type",
			buffer:       []byte{0x80}, // Accessory decoder (not implemented)
			opMode:       OperationsMode,
			svcModeReady: false,
			expectedType: UnknownMsg,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := &Decoder{
				opMode:       tt.opMode,
				svcModeReady: tt.svcModeReady,
			}
			msg := NewMessage(nil, decoder)
			msg.AddBytes(tt.buffer)

			if got := msg.messageType(); got != tt.expectedType {
				t.Errorf("messageType() = %v, want %v", got, tt.expectedType)
			}
		})
	}
}

func TestCheckAddress(t *testing.T) {
	tests := []struct {
		name          string
		messageBytes  []byte
		decoderAddr   []byte
		consistAddr   []byte
		snoop         bool
		expectedAddr  AddressType
		expectedMatch bool
	}{
		{
			name:          "Broadcast address",
			messageBytes:  []byte{0x00},
			decoderAddr:   []byte{0x03},
			consistAddr:   []byte{},
			snoop:         false,
			expectedAddr:  BroadcastAddress,
			expectedMatch: true,
		},
		{
			name:          "Idle packet",
			messageBytes:  []byte{0xFF},
			decoderAddr:   []byte{0x03},
			consistAddr:   []byte{},
			snoop:         false,
			expectedAddr:  IdleAddress,
			expectedMatch: false,
		},
		{
			name:          "Direct address match",
			messageBytes:  []byte{0x03, 0x01},
			decoderAddr:   []byte{0x03},
			consistAddr:   []byte{0x04},
			snoop:         false,
			expectedAddr:  DirectAddress,
			expectedMatch: true,
		},
		{
			name:          "Consist address match",
			messageBytes:  []byte{0x04, 0x01},
			decoderAddr:   []byte{0x03},
			consistAddr:   []byte{0x04},
			snoop:         false,
			expectedAddr:  ConsistAddress,
			expectedMatch: true,
		},
		{
			name:          "Unknown address with snoop",
			messageBytes:  []byte{0x05, 0x01},
			decoderAddr:   []byte{0x03},
			consistAddr:   []byte{0x04},
			snoop:         true,
			expectedAddr:  UnknownAddress,
			expectedMatch: true,
		},
		{
			name:          "Unknown address without snoop",
			messageBytes:  []byte{0x05, 0x01},
			decoderAddr:   []byte{0x03},
			consistAddr:   []byte{0x04},
			snoop:         false,
			expectedAddr:  UnknownAddress,
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := &Decoder{
				address:        tt.decoderAddr,
				consistAddress: tt.consistAddr,
				Snoop:          tt.snoop,
			}
			msg := NewMessage(nil, decoder)
			msg.AddBytes(tt.messageBytes)

			result := msg.checkAddress()

			if result != tt.expectedMatch {
				t.Errorf("checkAddress() = %v, want %v", result, tt.expectedMatch)
			}

			if msg.addr != tt.expectedAddr {
				t.Errorf("address type = %v, want %v", msg.addr, tt.expectedAddr)
			}
		})
	}
}

// TestMotionCommand_SpeedMode14 tests the motionCommand function against the SpeedMode14 input/output map
func TestMotionCommand_SpeedMode14(t *testing.T) {
	// Inputs given in reverse mode, add 0x20 for forward mode
	SpeedMode14 := map[byte]uint8{
		0x40: 0,  // 01000000 Stop
		0x41: 1,  // 01000001 Emergency Stop
		0x42: 2,  // 01000010 Speed 1
		0x43: 3,  // 01000011 Speed 2
		0x44: 4,  // 01000100 Speed 3
		0x45: 5,  // 01000101 Speed 4
		0x46: 6,  // 01000110 Speed 5
		0x47: 7,  // 01000111 Speed 6
		0x48: 8,  // 01001000 Speed 7
		0x49: 9,  // 01001001 Speed 8
		0x4A: 10, // 01001010 Speed 9
		0x4B: 11, // 01001011 Speed 10
		0x4C: 12, // 01001100 Speed 11
		0x4D: 13, // 01001101 Speed 12
		0x4E: 14, // 01001110 Speed 13
		0x4F: 15, // 01001111 Speed 14
	}

	decoder := &Decoder{
		speedMode: motor.SpeedMode14,
	}
	msg := NewMessage(nil, decoder)
	b := make([]byte, 1)

	for input, expected := range SpeedMode14 {
		// Test forward motion
		b[0] = input | 0x20 // Set direction bit to forward
		speed, reverse, ok := msg.motionCommand(b)

		if speed != expected {
			t.Errorf("motionCommand() SpeedMode14Fwd got %d, want %d", speed, expected)
		}
		if reverse {
			t.Errorf("motionCommand() SpeedMode14Fwd got reverse %t, want false", reverse)
		}
		if !ok {
			t.Errorf("motionCommand() SpeedMode14Fwd got ok %t, want true", ok)
		}

		// Test reverse motion
		b[0] = input
		speed, reverse, ok = msg.motionCommand(b)
		if speed != expected {
			t.Errorf("motionCommand() SpeedMode14Rev got %d, want %d", speed, expected)
		}
		if !reverse {
			t.Errorf("motionCommand() SpeedMode14Rev got reverse %t, want true", reverse)
		}
		if !ok {
			t.Errorf("motionCommand() SpeedMode14Rev got ok %t, want true", ok)
		}
	}
}

// TestMotionCommand_SpeedMode28 tests the motionCommand function against the SpeedMode28 input/output map
func TestMotionCommand_SpeedMode28(t *testing.T) {
	// Inputs given in reverse mode, add 0x20 for forward mode
	SpeedMode28 := map[byte]uint8{
		0x40: 0,  // 00000 Stop
		0x50: 0,  // 10000 Stop
		0x41: 1,  // 00001 Emergency Stop
		0x51: 1,  // 10001 Emergency Stop
		0x42: 2,  // 00010 Step 1
		0x52: 3,  // 10010 Step 2
		0x43: 4,  // 00011 Step 3
		0x53: 5,  // 10011 Step 4
		0x44: 6,  // 00100 Step 5
		0x54: 7,  // 10100 Step 6
		0x45: 8,  // 00101 Step 7
		0x55: 9,  // 10101 Step 8
		0x46: 10, // 00110 Step 9
		0x56: 11, // 10110 Step 10
		0x47: 12, // 00111 Step 11
		0x57: 13, // 10111 Step 12
		0x48: 14, // 01000 Step 13
		0x58: 15, // 11000 Step 14
		0x49: 16, // 01001 Step 15
		0x59: 17, // 11001 Step 16
		0x4A: 18, // 01010 Step 17
		0x5A: 19, // 11010 Step 18
		0x4B: 20, // 01011 Step 19
		0x5B: 21, // 11011 Step 20
		0x4C: 22, // 01100 Step 21
		0x5C: 23, // 11100 Step 22
		0x4D: 24, // 01101 Step 23
		0x5D: 25, // 11101 Step 24
		0x4E: 26, // 01110 Step 25
		0x5E: 27, // 11110 Step 26
		0x4F: 28, // 01111 Step 27
		0x5F: 29, // 11111 Step 28
	}

	decoder := &Decoder{
		speedMode: motor.SpeedMode28,
	}
	msg := NewMessage(nil, decoder)
	b := make([]byte, 1)

	for input, expected := range SpeedMode28 {
		// Test forward motion
		b[0] = input | 0x20 // Set direction bit to forward
		speed, reverse, ok := msg.motionCommand(b)

		if speed != expected {
			t.Errorf("motionCommand() SpeedMode28Fwd got %d, want %d", speed, expected)
		}
		if reverse {
			t.Errorf("motionCommand() SpeedMode28Fwd got reverse %t, want false", reverse)
		}
		if !ok {
			t.Errorf("motionCommand() SpeedMode28Fwd got ok %t, want true", ok)
		}

		// Test reverse motion
		b[0] = input
		speed, reverse, ok = msg.motionCommand(b)
		if speed != expected {
			t.Errorf("motionCommand() SpeedMode28Rev got %d, want %d", speed, expected)
		}
		if !reverse {
			t.Errorf("motionCommand() SpeedMode28Rev got reverse %t, want true", reverse)
		}
		if !ok {
			t.Errorf("motionCommand() SpeedMode28Rev got ok %t, want true", ok)
		}
	}
}

// TestMotionCommand_SpeedMode128 tests the motionCommand function
func TestMotionCommand_SpeedMode128(t *testing.T) {
	// There's no need for a map here, the function should just return the speed byte
	decoder := &Decoder{
		speedMode: motor.SpeedMode128,
	}
	msg := NewMessage(nil, decoder)
	b := make([]byte, 2)

	for i := uint8(0); i < 128; i++ {
		// Test forward motion
		b[0] = 0x3F     // 00111111
		b[1] = i | 0x80 // Set bit 7 for forward
		speed, reverse, ok := msg.motionCommand(b)

		if speed != i {
			t.Errorf("motionCommand() SpeedMode128 got %d, want %d", speed, i)
		}
		if reverse {
			t.Errorf("motionCommand() SpeedMode128 got reverse %t, want false", reverse)
		}
		if !ok {
			t.Errorf("motionCommand() SpeedMode128 got ok %t, want true", ok)
		}

		// Test reverse motion
		b[0] = 0x3F // 00111111
		b[1] = i
		speed, reverse, ok = msg.motionCommand(b)

		if speed != i {
			t.Errorf("motionCommand() SpeedMode128 got %d, want %d", speed, i)
		}
		if !reverse {
			t.Errorf("motionCommand() SpeedMode128 got reverse %t, want true", reverse)
		}
		if !ok {
			t.Errorf("motionCommand() SpeedMode128 got ok %t, want true", ok)
		}
	}
}
