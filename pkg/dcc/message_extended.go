package dcc

func (m *Message) extendedPacket() bool {
	// Extended message format
	l := len(m.buf)

	// First command byte
	fb := 1
	switch m.addr {
	case BroadcastAddress:
		// Broadcast address is always 1 byte, no change
	case DirectAddress:
		fb = len(m.decoder.address)
	case ConsistAddress:
		fb = len(m.decoder.consistAddress)
	}

	// Command type is in the top 3 bits of the first command byte
	switch m.buf[fb] >> 5 {
	case 0b000:
		return m.decoderConsistControlInstruction(m.buf[fb : l-1])
	case 0b001:
		return m.advancedOperationInstruction(m.buf[fb : l-1])
	case 0b010, 0b011:
		// Speed and Direction Instruction
		speed, reverse, ok := m.motionCommand(m.buf[fb : l-1])
		if ok {
			m.decoder.motor.SetSpeed(speed, reverse)
		}
		return ok
	case 0b100:
		return m.functionGroupOneInstruction(m.buf[fb])
	case 0b101:
		return m.functionGroupTwoInstruction(m.buf[fb])
	case 0b110:
		return m.featureExpansion(m.buf[fb : l-1])
	case 0b111:
		// Don't allow editing CVs from consist-addressed messages
		if m.addr == ConsistAddress {
			return false
		}
		return m.configVariableAccessInstruction(m.buf[fb : l-1])
	}
	return false
}

func (m *Message) decoderConsistControlInstruction(b []byte) bool {
	l := len(b)
	switch b[0] >> 4 {
	// 0b0000xxxx
	case 0b0000:
		// Decoder control 0000xxxx
		switch b[0] {
		case 0b00000000:
			if l == 1 && m.addr == BroadcastAddress {
				println("reset packet received")
				m.decoder.Reset()
				return true
			}
		case 0b00000001:
			// Decoder hard reset packet
			m.cv.Reset(29)
			m.cv.Reset(31)  // CV257-512 index
			m.cv.Reset(32)  // CV257-512 index
			m.cv.Set(19, 0) // Consist address
			m.decoder.Reset()
			return true
		case 0b00000010, 0b00000011:
			// Factory test mode, not supported
		case 0b00001010, 0b00001011:
			// Set advanced addressing mode (CV29 bit 5)
			cv29 := m.cv.CV(29)
			return m.cv.Set(29, (cv29&^0b00100000)|(b[0]&1)<<5)
		case 0b00001111:
			// Decoder ack request
			return true
		}
	// 0b0001xxxx
	case 0b0001:
		/* Consist control 0001xxxx
		TODO: Implement the other logic:
		When Consist Control is in effect, the decoder will ignore any speed or direction instructions
		addressed to its normal locomotive address (unless this address is the same as its consist address).
		190 Speed and direction instructions now apply to the consist address only

		Functions controlled by Function Group One (100) and Function Group Two (101) will continue to
		respond to the decoder’s baseline address. Functions controlled by instructions 100 and 101 also
		respond to the consist address if the appropriate bits in CVs 21 and 22 have been activated.

		By default, all forms of Bi-directional communication are not activated in response to commands
		sent to the consist address until specifically activated by a Decoder Control instruction.

		https://www.nmra.org/sites/default/files/standards/sandrp/DCC/S/s-9.2.1_dcc_extended_packet_formats.pdf
		Page 5
		*/
		switch b[0] {
		case 0b00010010:
			// Set consist address
			if l > 1 && b[1] < 128 {
				return m.cv.Set(19, b[1])
			}
		case 0b00010011:
			// Set consist address and reverse direction
			if l > 1 {
				return m.cv.Set(19, b[1]|0x80)
			}
		}
	}
	return false
}

func (m *Message) advancedOperationInstruction(b []byte) bool {
	// 0b001xxxxx
	// 128-step speed control
	switch b[0] {
	case 0b00111111:
		speed, reverse, ok := m.motionCommand(b)
		if ok {
			m.decoder.motor.SetSpeed(speed, reverse)
		}
		return ok
	default:
		return false
	}
}

func (m *Message) functionGroupOneInstruction(b uint8) bool {
	// Function Group One Instruction
	// 100DDDDD
	// If message was sent to the consist address, ignore function values according to CVs 21 and 22
	if m.addr == ConsistAddress {
		b &= m.decoder.consistFuncMask[0]
	}
	// FL (F0)
	m.decoder.callFunction(0, b&(1<<4) != 0)
	// F1-F4
	for i := range uint16(4) {
		m.decoder.callFunction(i+1, b&(1<<i) != 0)
	}
	return true
}

func (m *Message) functionGroupTwoInstruction(b uint8) bool {
	// Function Group Two Instruction
	// 101SDDDD
	// Bit 4 (S) is a shift bit. If set to 1, bits 0-3 (D) are F5-F8. If set to 0, bits 0-3 are F9-F12
	offset := uint16(5)
	mask := m.decoder.consistFuncMask[1]
	if b&(1<<4) == 0 {
		offset = 9
		mask = m.decoder.consistFuncMask[2]
	}
	// If message was sent to the consist address, ignore function values according to CVs 21 and 22
	if m.addr == ConsistAddress {
		b &= mask
	}
	for i := range uint16(4) {
		m.decoder.callFunction(i+offset, b&(1<<i) != 0)
	}
	return true
}

func (m *Message) functionGroupNInstruction(n uint16, b uint8) bool {
	// Function Group N Instruction
	// Each bit in the command byte represents a function (starting with F13)
	for i := range uint16(8) {
		m.decoder.callFunction(i+n, b&(1<<i) != 0)
	}
	return true
}

func (m *Message) featureExpansion(b []byte) bool {
	// Feature Expansion Instruction
	// 110GGGGG DDDDDDDD [DDDDDDDD]
	switch b[0] {
	case 0b11000000:
	// Binary state control long form
	// 32,767 binary states, 0 resets all states to off
	// 11000000 DLLLLLLL HHHHHHHH
	// L = 7-bit low byte
	// H = optional 8-bit high byte (treated as 0 if not present)
	// D = data bit
	// TODO: Implement
	case 0b11000001:
	// Model time and date command
	// Not supported at this time, possibly in the future
	case 0b11000010:
	// System time (0-65535 milliseconds)
	// Not supported at this time, possibly in the future
	case 0b11011101:
	// Binary state control short form
	// 127 binary states, 0 resets all states to off
	// 11011101 DLLLLLLL
	// L = 7-bit low byte
	// D = data bit
	// TODO: Implement
	case 0b11011110:
		// Functions F13-F20
		return m.functionGroupNInstruction(13, b[1])
	case 0b11011111:
		// Functions F21-F28
		return m.functionGroupNInstruction(21, b[1])
	case 0b11011000:
		// Functions F29-F36
		return m.functionGroupNInstruction(29, b[1])
	case 0b11011001:
		// Functions F37-F44
		return m.functionGroupNInstruction(37, b[1])
	case 0b11011010:
		// Functions F45-F52
		return m.functionGroupNInstruction(45, b[1])
	case 0b11011011:
		// Functions F53-F60
		return m.functionGroupNInstruction(53, b[1])
	case 0b11011100:
		// Functions F61-F68
		return m.functionGroupNInstruction(61, b[1])
	}
	return false
}

func (m *Message) configVariableAccessInstruction(b []byte) bool {
	l := len(b)
	// All CV access commands are at least 2 bytes
	if l < 2 {
		return false
	}
	// Long form - 1110xxxx
	if b[0]&0xF0 == 0xE0 {
		if l == 3 {
			// Format is 1110CCAA AAAAAAAA DDDDDDDD
			return m.cvCommand(m.cv.IndexPage(), b)
		} else if l > 3 {
			// XPOM - Extended Programming On Main
			// Up to 8 bytes plus short/long address and checksum (max 11 bytes)
			// 1110GGSS VVVVVVVV VVVVVVVV VVVVVVVV [DDDDDDDD [DDDDDDDD [DDDDDDDD [DDDDDDDD]]]]
			return m.handleXPOM(b)
		}
	} else {
		// Short form - 1111xxxx
		switch b[0] {
		case 0b11110010:
			// CV23 Acceleration rate adjustment
			return m.cv.Set(23, b[1])
		case 0b11110011:
			// CV24 Deceleration rate adjustment
			return m.cv.Set(24, b[1])
		case 0b11110100:
			// Extended address programming (CV17, CV18, CV29)
			// Must receive two identical packets to confirm before setting
			if l < 3 {
				return false
			}
			// Check for confirmed values
			if m.cvConfirmCheck(17, b[1]) && m.cvConfirmCheck(18, b[2]) {
				// Values confirmed, set the CVs
				cv29 := m.cv.CV(29) &^ (1 << 5) // Clear bit 5
				if b[1] >= 192 && b[1] <= 231 {
					// If CV17 is valid, set bit 5 to enable it
					cv29 |= 1 << 5
				}
				return m.cv.Set(17, b[1]) &&
					m.cv.Set(18, b[2]) &&
					m.cv.Set(29, cv29)
			} else {
				// cvConfirmCheck will store the values for the next packet
				return true
			}
		case 0b11110101:
			// Indexed CVs (CV257-512), set CV31/32 index values
			// Must receive two identical packets to confirm before setting
			if l < 3 || b[1] < 16 {
				// CV31 values < 16 are reserved
				return false
			}
			// Check for confirmed values
			if m.cvConfirmCheck(31, b[1]) && m.cvConfirmCheck(32, b[2]) {
				// Values confirmed, load the new index
				err := m.cv.LoadIndex(b[1], b[2])
				if err != nil {
					println("could not load index: " + err.Error())
					return false
				}
				return true
			} else {
				// cvConfirmCheck will store the values for the next packet
				return true
			}
		case 0b11110110:
			// Consist extended address
			// Must receive two identical packets to confirm before setting
			if l < 3 {
				return false
			}
			// Check for confirmed values
			if m.cvConfirmCheck(19, b[1]) && m.cvConfirmCheck(20, b[2]) {
				// Values confirmed, set the CVs
				return m.cv.Set(19, b[1]) && m.cv.Set(20, b[2])
			} else {
				// cvConfirmCheck will store the values for the next packet
				return true
			}
		case 0b11111001:
			// Service Mode Decoder Lock S-9.2.3 Appendix B
			// Not implemented
		}
	}

	return false
}

// Confirm we've received the same value for a CV twice before setting it
func (m *Message) cvConfirmCheck(cv uint16, value uint8) bool {
	value, ok := m.cvConfirm[cv]
	if ok && value == value {
		return true
	}
	m.cvConfirm[cv] = value
	return false
}
