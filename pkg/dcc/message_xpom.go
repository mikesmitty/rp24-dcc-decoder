package dcc

func (m *Message) handleXPOM(b []byte) bool {
	// XPOM - Extended Programming On Main
	// Up to 8 bytes plus short/long address and checksum (max 11 bytes)
	// Two identical packets must be received to confirm correctness

	// All XPOM commands are at least 4 bytes
	if len(b) < 4 {
		return false
	}

	// Don't process XPOM commands addressed to the consist address
	if m.addr == ConsistAddress {
		return false
	}

	// Check for identical packets
	match := false
	if len(b) == len(m.lastXPOM) {
		match = true
		for i, v := range b {
			if v != m.lastXPOM[i] {
				match = false
				break
			}
		}
	}
	// If we don't have a match, store this one and wait for the next
	// Technically any addressed messages received including broadcasts should invalidate
	// lastXPOM but for now we'll just make sure the last received XPOM message is identical
	if !match {
		m.lastXPOM = m.lastXPOM[:0]
		for _, v := range b {
			m.lastXPOM = append(m.lastXPOM, v)
		}
		return false
	}

	// 1110CCSS AAAAAAAA AAAAAAAA AAAAAAAA [DDDDDDDD [DDDDDDDD [DDDDDDDD [DDDDDDDD]]]]
	// CC = Command
	// SS = Sequence number
	// A = CV Number (CV31 index, CV32 index, Relative CV Number)
	// D = Data

	// Commands:
	// 01 = read bytes
	// 11 = write bytes
	// 10 = write bits

	// seq := b[0] & 0b11 // FIXME: What is this for?
	// Find in https://www.nmra.org/sites/default/files/standards/sandrp/Draft/DCC/s-9.3.2_bi-directional_communication.pdf
	index := m.cv.IndexPage(b[1], b[2])
	switch (b[0] >> 2) & 0b11 {
	case 0b01:
		// Read bytes
		// All XPOM commands respond with four consecutive CV values
		// This one just doesn't write anything
	case 0b11:
		// Write bytes
		for i, v := range b[4:] {
			if !m.cv.IndexedSet(index, uint16(b[3])+uint16(i), v) {
				return false
			}
		}
	case 0b10:
		// Write bits
		if len(b) < 5 {
			return false
		}
		m.setCVCommand(index, b[3:])
	}

	// Return the four bytes via BiDi communication
	for i := uint16(0); i < 4; i++ {
		// FIXME: Do BiDi communication
		m.cv.IndexedCV(index, uint16(b[3])+i)
	}

	return true
}
