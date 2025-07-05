package dcc

func (m *Message) serviceModePacket(b []byte) bool {
	m.decoder.SetOpMode(ServiceMode)
	/* TODO: Implement the ability to leave service mode
	For now, service mode is hotel california. We'll depend on losing power to return to operations mode
	if time.Now().Sub(m.decoder.lastSvcResetTime) > 20*time.Millisecond {
		// If it's been more than 20ms since the last reset packet, service mode is no longer ready, back to operations mode
		// Check to ensure we didn't receive a service mode packet before switching back to operations mode per S-9.2.3C
		//m.decoder.SetOpMode(UndefinedMode)
		return
	} */

	ack := false

	// Direct CV Addressing commands
	// 0111CCAA AAAAAAAA DDDDDDDD EEEEEEEE
	if b[0]&0b11110000 == 0b01110000 {
		// CV access/programming command
		ack = m.cvCommand(m.cv.IndexPage(), b)
	}

	return ack
}

func (m *Message) cvCommand(index uint16, b []byte) bool {
	// CV programming command format
	// C = command type, A = address, D = data
	// C: 01 = verify, 11 = write byte
	// xxxxCCAA AAAAAAAA DDDDDDDD
	//
	// C: 10 = bit manipulation
	// F: 0 = verify, 1 = write
	// D: bit value
	// B: bit position
	// xxxxCCAA AAAAAAAA 111FDBBB

	op := (b[0] >> 2) & 0b11                         // CC
	cvNum := uint16(b[0]&0b11)<<8 | uint16(b[1]) + 1 // AA AAAAAAAA (n-1, CV1 = 0)
	data := b[2]                                     // DDDDDDDD

	ack := false
	switch op {
	case 0b01:
		// Verify byte
		if v, ok := m.cv.IndexedCVOk(index, cvNum); ok && v == data {
			ack = true
		}
	case 0b10:
		// Bit manipulation (write/verify bit)
		if data&0b11100000 != 0b11100000 {
			// Invalid command
			return false
		}
		pos := data & 0b111
		bit := (data >> 3) & 1
		v, ok := m.cv.IndexedCVOk(index, cvNum)

		if data&0b10000 == 0 {
			// Verify
			if ok && (v>>pos)&1 == bit {
				ack = true
			}
		} else if ok {
			// Write
			if ok := m.cv.IndexedSetSync(index, cvNum, (v&^(1<<pos))|(bit<<pos)); ok {
				ack = true
			} else {
				println("CV write error")
			}
		} else {
			println("CV not found")
		}
	case 0b11:
		// Write byte
		if ok := m.cv.IndexedSetSync(index, cvNum, data); ok {
			ack = true
		} else {
			println("CV write error")
		}
	}
	return ack
}
