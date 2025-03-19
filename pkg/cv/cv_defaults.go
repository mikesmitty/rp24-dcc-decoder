package cv

import "github.com/mikesmitty/rp24-dcc-decoder/pkg/store"

func (c *CVHandler) setDefaults(s *store.Store, version uint8) {
	roPersist := store.Persistent | store.ReadOnly

	// Number, Default, Flags
	s.SetDefault(1, 3, store.Persistent)   // ADDR: Primary address
	s.SetDefault(2, 10, store.Persistent)  // MOTOR: Vstart (minimum throttle required to start moving)
	s.SetDefault(3, 0, store.Persistent)   // MOTOR: Default acceleration rate (0 = immediate)
	s.SetDefault(4, 0, store.Persistent)   // MOTOR: Default deceleration rate (0 = immediate)
	s.SetDefault(5, 255, store.Persistent) // MOTOR: Vmax - maximum voltage
	s.SetDefault(6, 128, store.Persistent) // MOTOR: Vmid - mid-range voltage
	s.SetDefault(7, version, roPersist)    // SYS: Version number
	s.SetDefault(8, 0x0D, store.ReadOnly)  // SYS: Manufacturer ID: "Public Domain & DIY Decoders"
	s.SetDefault(9, 40, store.Persistent)  // MOTOR: PWM frequency in kHz (1-250)
	s.SetDefault(10, 0, store.Persistent)  // MOTOR: Back EMF motor control cutoff speed
	s.SetDefault(11, 10, store.Persistent) // SYS: Control packet keepalive timeout in 100ms units

	// Extended address - Top 2 bits of MSB must be 1 and are ignored (min 192, max 231), allowing for any 4 digit number
	s.SetDefault(17, 0, store.Persistent) // ADDR: MSB
	s.SetDefault(18, 0, store.Persistent) // ADDR: LSB
	s.SetDefault(19, 0, store.Persistent) // CONSIST: ADDR: Consist address
	s.SetDefault(20, 0, store.Persistent) // CONSIST: ADDR: Consist extended address

	s.SetDefault(23, 0, store.Persistent) // MOTOR: No acceleration adjustment
	s.SetDefault(24, 0, store.Persistent) // MOTOR: No deceleration adjustment
	s.SetDefault(25, 2, store.Persistent) // MOTOR: Linear speed curve by default FIXME: Check on this

	// CV 29:
	// Bit 7: 0 = Mobile decoder, 1 = Accessory decoder
	// Bit 6: Reserved
	// Bit 5: 0 = Short address mode, 1 = Extended address mode
	// Bit 4: 0 = CV 2,5,6 speed curve, 1 = CV 25 speed table
	// Bit 3: 1 = RailCom enabled
	// Bit 2: 0 = DCC only, 1 = DCC & DC
	// Bit 1: 0 = 14 speed steps, 1 = 28/128 speed steps
	// Bit 0: 0 = Forward direction, 1 = Reverse direction
	s.SetDefault(29, 0b00001010, store.Persistent) // RailCom enabled, 28/128 speed steps

	s.SetDefault(31, 0, store.Volatile) // INDEX: CV index paging MSB (0 is disabled, 1-15 are reserved)
	s.SetDefault(32, 0, store.Volatile) // INDEX: CV index paging LSB
	// CV 33-46 are reserved for function mapping, but in a rather restrictive way. Will implement another method elsewhere
	s.SetDefault(51, 10, store.Persistent)  // MOTOR: Low to high PID gain cutover speed step
	s.SetDefault(52, 10, store.Persistent)  // MOTOR: Low speed Kp gain (proportional)
	s.SetDefault(53, 130, store.Persistent) // MOTOR: Max speed EMF voltage
	s.SetDefault(54, 50, store.Persistent)  // MOTOR: High speed Kp gain (proportional)
	s.SetDefault(55, 100, store.Persistent) // MOTOR: Ki gain (integral)
	s.SetDefault(56, 255, store.Persistent) // MOTOR: Low speed PID scaling factor

	s.SetDefault(116, 50, store.Persistent)  // MOTOR: Speed step 1 back EMF measurement interval in 0.1ms steps (50-200)
	s.SetDefault(117, 150, store.Persistent) // MOTOR: Speed step max back EMF measurement interval in 0.1ms steps (50-200)
	s.SetDefault(118, 15, store.Persistent)  // MOTOR: Speed step 1 back EMF measurement cutout duration in 0.1ms steps (10-40)
	s.SetDefault(119, 20, store.Persistent)  // MOTOR: Speed step max back EMF measurement cutout duration in 0.1ms steps (10-40)

	/* FIXME: Cleanup. This would require AUX numbers to be n+2 which is annoying (AUX1/output 3, etc.)
	// Function mapping
	s.SetDefault(33, 0b00000001, store.Persistent) // F0f
	s.SetDefault(34, 0b00000010, store.Persistent) // F0r
	s.SetDefault(35, 0b00000100, store.Persistent) // F1
	s.SetDefault(36, 0b00001000, store.Persistent) // F2
	s.SetDefault(37, 0b00010000, store.Persistent) // F3
	s.SetDefault(38, 0b00000100, store.Persistent) // F4
	s.SetDefault(39, 0b00001000, store.Persistent) // F5
	s.SetDefault(40, 0b00010000, store.Persistent) // F6
	s.SetDefault(41, 0b00100000, store.Persistent) // F7
	s.SetDefault(42, 0b01000000, store.Persistent) // F8
	s.SetDefault(43, 0b00010000, store.Persistent) // F9
	s.SetDefault(44, 0b00100000, store.Persistent) // F10
	s.SetDefault(45, 0b01000000, store.Persistent) // F11
	s.SetDefault(46, 0b10000000, store.Persistent) // F12
	*/
}
