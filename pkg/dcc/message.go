package dcc

import (
	"sync"

	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/motor"
)

type MessageType int

const (
	UnknownMsg MessageType = iota
	ServiceMsg
	ExtendedMsg
	AdvancedExtendedMsg
)

type AddressType int

const (
	UnknownAddress AddressType = iota
	BroadcastAddress
	DirectAddress
	ConsistAddress
	IdleAddress
)

type Message struct {
	addr    AddressType
	buf     []byte
	msgType MessageType
	mutex   sync.Mutex

	cv      cv.Handler
	decoder *Decoder

	cvConfirm map[uint16]uint8
	lastXPOM  []byte
}

func NewMessage(cvHandler cv.Handler, decoder *Decoder) *Message {
	return &Message{
		buf:       make([]byte, 0, maxMsgLength),
		cv:        cvHandler,
		cvConfirm: make(map[uint16]uint8),
		decoder:   decoder,
		lastXPOM:  make([]byte, 0, 8),
	}
}

func (m *Message) AddByte(b byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.buf = append(m.buf, b)
}

func (m *Message) AddBytes(b []byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.buf = append(m.buf, b...)
}

func (m *Message) Bytes() []byte {
	return m.buf
}

func (m *Message) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.buf = m.buf[:0]
	m.msgType = UnknownMsg
}

func (m *Message) XOR() bool {
	var xor byte
	for _, b := range m.buf {
		xor ^= b
	}
	return xor == 0
}

func (m *Message) IsEmpty() bool {
	return len(m.buf) == 0
}

func (m *Message) IsFull() bool {
	return len(m.buf) == cap(m.buf)
}

func (m *Message) Length() int {
	return len(m.buf)
}

func (m *Message) Process() {
	if len(m.buf) == 0 {
		// No message received yet, nothing to process
		return
	}

	// Make sure we don't overlap message processing
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.decoder.Snoop && (!m.checkAddress() || m.addr == UnknownAddress) {
		// Ignore messages not addressed to us if we're not being nosy
		return
	}

	// Check the message type to determine how to handle it
	m.msgType = m.messageType()

	ok := false
	switch m.msgType {
	case ServiceMsg:
		ok = m.serviceModePacket(m.buf)
	case ExtendedMsg:
		m.extendedPacket(m.buf)
	case AdvancedExtendedMsg:
		m.advancedExtendedPacket(m.buf)
	default:
		// Unknown message type
	}

	if ok {
		m.decoder.BasicAck()
		// FIXME: Advanced Ack usage?
		// m.decoder.AdvancedAck([]byte{ackByte}, []byte{})
	}

	// respByte := nackByte
	// if ok {
	// 	respByte = ackByte
	// }
	// err := m.decoder.AdvancedAck([]byte{respByte}, []byte{})
	// if err != nil {
	// 	// println("Error sending ack:", err.Error()) // FIXME: Cleanup
	// }
}

func (m *Message) messageType() MessageType {
	if len(m.buf) == 0 {
		// No message received yet, return unknown
		return UnknownMsg
	}

	b := m.buf[0]

	if m.decoder.opMode == ServiceMode || m.decoder.svcModeReady {
		if b >= 112 && b <= 127 {
			// Received a service mode packet
			return ServiceMsg
		} else {
			// Ignore non-service mode packets while in service mode
			return UnknownMsg
		}
	}

	if b < 128 || b >= 192 && b <= 231 {
		// Extended message format
		return ExtendedMsg
	} else if b == 253 || b == 254 {
		// Advanced extended message format
		return AdvancedExtendedMsg
	}
	// We're intentionally ignoring accessory decoder messages for now, may implement later
	// Basic messages are superseded by extended messages
	return UnknownMsg
}

func (m *Message) motionCommand(bytes []byte) (uint8, bool, bool) {
	if len(bytes) > 1 && m.decoder.motor.SpeedMode() == motor.SpeedMode14 {
		// We're configured for 14 speed mode, ignore 128 speed commands
		return 1, false, false
	}

	var speed uint8
	var reverse bool

	switch len(bytes) {
	case 1:
		speed = uint8(bytes[0] & 0x0F)
		reverse = (bytes[0] & 0b00100000) == 0
		if m.decoder.motor.SpeedMode() == motor.SpeedMode14 {
			// 14-speed mode: 01DCSSSS
			// D: Direction (0 = reverse, 1 = forward)
			// S: Speed step (0-15)
			// C: Ignored
			// No changes needed
		} else {
			// 28-speed mode: 01DLSSSS
			// D: Direction (0 = reverse, 1 = forward)
			// L: Low bit
			// S: Speed step (0-27)
			speed = speed<<1 | (bytes[0] >> 4 & 1)
			if speed < 4 {
				// Stop and emergency stop ignore the extra low bit value
				speed = bytes[0] & 0x01
			} else {
				speed -= 2
			}
			// 28/128 speed modes are selected by the last speed command received
			m.decoder.motor.SetSpeedMode(motor.SpeedMode28)
		}
	case 2:
		if bytes[0] == 0b00111111 {
			// 128-speed mode: 00111111 DSSSSSSS
			// D: Direction (0 = reverse, 1 = forward)
			// S: Speed step (0-127)
			speed = bytes[1] & 0x7F
			reverse = (bytes[1] & 0x80) == 0

			// 28/128 speed modes are selected by the last speed command received
			m.decoder.motor.SetSpeedMode(motor.SpeedMode128)
		} else {
			// Invalid speed mode
			return 1, false, false
		}
	default:
		println("Invalid speed command length")
		return 1, false, false
	}

	return speed, reverse, true
}

func (m *Message) checkAddress() bool {
	if len(m.buf) == 0 {
		// No message received yet, return invalid
		return false
	}

	switch m.buf[0] {
	case 0x00:
		// Broadcast address
		m.addr = BroadcastAddress
		return true
	case 0xFF:
		// Idle packet, ignore
		m.addr = IdleAddress
		// TODO: Idle packets count as valid data packets to return to operations mode, I think?
	case 253, 254:
		// Advanced extended packet format, not supported yet
	default:
		// We don't know how to interprest accessory decoder messages
		if m.buf[0] >= 128 && m.buf[0] <= 191 {
			m.addr = UnknownAddress
			return false
		}

		// Check for direct or consist address match
		if m.addressMatch(m.decoder.address) {
			m.addr = DirectAddress
			return true
		} else if m.addressMatch(m.decoder.consistAddress) {
			m.addr = ConsistAddress
			return true
		}

		// TODO: Differentiate recognized message types with unrecognized addresses while snooping?
		// TODO: Remove snooping once address setting is done being implemented
		// No recognized addresses
		m.addr = UnknownAddress
		// If we're snooping that's okay anyway
		return m.decoder.Snoop
	}
	return false
}

// addressMatch checks if the first bytes of the message match the provided buffer
func (m *Message) addressMatch(b []byte) bool {
	if len(b) > len(m.buf) {
		return false
	}
	for i, v := range b {
		if m.buf[i] != v {
			return false
		}
	}
	return true
}
