package cv

import (
	"errors"

	"github.com/mikesmitty/rp24-dcc-decoder/pkg/store"
)

// Save the current version to flash in case we decide to change the storage format or update logic
const roPersist = store.Persistent | store.ReadOnly

const maxCVIndexPage = 0

func (c *CVHandler) LoadIndex(cv31, cv32 uint8) error {
	index := c.IndexPage(cv31, cv32)
	if index > maxCVIndexPage {
		return errors.New("invalid index page")
	}

	println("Loading CV index page", index)

	// Persist any remaining dirty flags
	if ok := c.cvStore.ProcessChanges(); !ok {
		// Going to not consider this a critical error at least for now
		println("could not save changes to flash")
	}

	// Clear out the existing CVs and load the new defaults
	c.cvStore.Clear()
	switch index {
	// case 0
	default:
		// Number, Default, Flags
		c.cvStore.SetDefault(1, 3, store.Persistent)     // ADDR: Primary address
		c.cvStore.SetDefault(2, 10, store.Persistent)    // MOTOR: Vstart (minimum throttle required to start moving)
		c.cvStore.SetDefault(3, 0, store.Persistent)     // MOTOR: Default acceleration rate (0 = immediate)
		c.cvStore.SetDefault(4, 0, store.Persistent)     // MOTOR: Default deceleration rate (0 = immediate)
		c.cvStore.SetDefault(5, 255, store.Persistent)   // MOTOR: Vmax - maximum voltage
		c.cvStore.SetDefault(6, 128, store.Persistent)   // MOTOR: Vmid - mid-range voltage
		c.cvStore.SetDefault(7, fwVersion[0], roPersist) // SYS: Major version number
		c.cvStore.SetDefault(8, 0x0D, store.ReadOnly)    // SYS: Manufacturer ID: "Public Domain & DIY Decoders"
		c.cvStore.SetDefault(9, 40, store.Persistent)    // MOTOR: PWM frequency in kHz (1-250)
		c.cvStore.SetDefault(10, 0, store.Persistent)    // MOTOR: Back EMF motor control cutoff speed
		c.cvStore.SetDefault(11, 10, store.Persistent)   // SYS: Control packet keepalive timeout in 100ms units

		// Extended address - Top 2 bits of MSB must be 1 and are ignored (min 192, max 231), allowing for any 4 digit number
		c.cvStore.SetDefault(17, 0, store.Persistent) // ADDR: MSB
		c.cvStore.SetDefault(18, 0, store.Persistent) // ADDR: LSB
		c.cvStore.SetDefault(19, 0, store.Persistent) // CONSIST: ADDR: Consist address
		c.cvStore.SetDefault(20, 0, store.Persistent) // CONSIST: ADDR: Consist extended address
		c.cvStore.SetDefault(21, 0, store.Persistent) // CONSIST: Consist address function activation F1-F8
		c.cvStore.SetDefault(22, 0, store.Persistent) // CONSIST: Consist address function activation F0f, F0r, F9-F12
		c.cvStore.SetDefault(23, 0, store.Persistent) // MOTOR: No acceleration adjustment
		c.cvStore.SetDefault(24, 0, store.Persistent) // MOTOR: No deceleration adjustment

		// CV 28:
		// Used to configure decoder’s Bi-Directional communication characteristics when CV29-Bit 3 is set
		// Bit 0 = Enable/Disable Unsolicited Decoder Initiated Transmission
		// Bit 1 = Enable/Disable Initiated Broadcast Transmission using Asymmetrical DCC Signal
		// Bit 2 = Enable/Disable Initiated Broadcast Transmission using Signal Controlled Influence Signal
		c.cvStore.SetDefault(28, 0, store.Persistent) // BiDi: Off

		// CV 29:
		// Bit 7: 0 = Mobile decoder, 1 = Accessory decoder
		// Bit 6: Reserved
		// Bit 5: 0 = Short address mode, 1 = Extended address mode
		// Bit 4: 0 = CV 2,5,6 speed curve, 1 = CV 25 speed table
		// Bit 3: 1 = RailCom enabled
		// Bit 2: 0 = DCC only, 1 = DCC & DC
		// Bit 1: 0 = 14 speed steps, 1 = 28/128 speed steps
		// Bit 0: 0 = Forward direction, 1 = Reverse direction
		c.cvStore.SetDefault(29, 0b00000010, store.Persistent) // BiDi disabled, 28/128 speed steps TODO: Enable BiDi

		c.cvStore.SetDefault(30, 0, store.Volatile) // ERROR: Error code TODO: Implement error codes
		c.cvStore.SetDefault(31, 0, store.Volatile) // INDEX: CV index paging MSB (0 is disabled, 1-15 are reserved)
		c.cvStore.SetDefault(32, 0, store.Volatile) // INDEX: CV index paging LSB

		// TODO: Reimplement for more flexibility
		c.cvStore.SetDefault(33, 0b00000001, store.Persistent) // FUNCTIONS: Output mapping for F0f
		c.cvStore.SetDefault(34, 0b00000010, store.Persistent) // FUNCTIONS: Output mapping for F0r
		c.cvStore.SetDefault(35, 0b00000100, store.Persistent) // FUNCTIONS: Output mapping for F1
		c.cvStore.SetDefault(36, 0b00001000, store.Persistent) // FUNCTIONS: Output mapping for F2
		c.cvStore.SetDefault(37, 0b00010000, store.Persistent) // FUNCTIONS: Output mapping for F3
		c.cvStore.SetDefault(38, 0b00000100, store.Persistent) // FUNCTIONS: Output mapping for F4
		c.cvStore.SetDefault(39, 0b00001000, store.Persistent) // FUNCTIONS: Output mapping for F5
		c.cvStore.SetDefault(40, 0b00010000, store.Persistent) // FUNCTIONS: Output mapping for F6
		c.cvStore.SetDefault(41, 0b00100000, store.Persistent) // FUNCTIONS: Output mapping for F7
		c.cvStore.SetDefault(42, 0b01000000, store.Persistent) // FUNCTIONS: Output mapping for F8
		c.cvStore.SetDefault(43, 0b00010000, store.Persistent) // FUNCTIONS: Output mapping for F9
		c.cvStore.SetDefault(44, 0b00100000, store.Persistent) // FUNCTIONS: Output mapping for F10
		c.cvStore.SetDefault(45, 0b01000000, store.Persistent) // FUNCTIONS: Output mapping for F11
		c.cvStore.SetDefault(46, 0b10000000, store.Persistent) // FUNCTIONS: Output mapping for F12

		c.cvStore.SetDefault(49, 1, store.Persistent)   // MOTOR: Enable back EMF motor control
		c.cvStore.SetDefault(50, 40, store.Persistent)  // MOTOR: Back EMF measurement settle delay in 5us steps
		c.cvStore.SetDefault(51, 10, store.Persistent)  // MOTOR: Low to high PID gain cutover speed step
		c.cvStore.SetDefault(52, 10, store.Persistent)  // MOTOR: Low speed Kp gain (proportional)
		c.cvStore.SetDefault(53, 130, store.Persistent) // MOTOR: Max speed EMF voltage
		c.cvStore.SetDefault(54, 50, store.Persistent)  // MOTOR: High speed Kp gain (proportional)
		c.cvStore.SetDefault(55, 100, store.Persistent) // MOTOR: Ki gain (integral)
		c.cvStore.SetDefault(56, 255, store.Persistent) // MOTOR: Low speed PID scaling factor

		c.cvStore.SetDefault(65, 0, store.Persistent)   // MOTOR: Startup kick to overcome static friction from a stop to speed step 1
		c.cvStore.SetDefault(66, 128, store.Persistent) // MOTOR: Forward trim
		// CV67-CV94: Speed table
		c.cvStore.SetDefault(67, 0, store.Persistent)   // MOTOR: Speed 1
		c.cvStore.SetDefault(68, 0, store.Persistent)   // MOTOR: Speed 2
		c.cvStore.SetDefault(69, 0, store.Persistent)   // MOTOR: Speed 3
		c.cvStore.SetDefault(70, 0, store.Persistent)   // MOTOR: Speed 4
		c.cvStore.SetDefault(71, 0, store.Persistent)   // MOTOR: Speed 5
		c.cvStore.SetDefault(72, 0, store.Persistent)   // MOTOR: Speed 6
		c.cvStore.SetDefault(73, 0, store.Persistent)   // MOTOR: Speed 7
		c.cvStore.SetDefault(74, 0, store.Persistent)   // MOTOR: Speed 8
		c.cvStore.SetDefault(75, 0, store.Persistent)   // MOTOR: Speed 9
		c.cvStore.SetDefault(76, 0, store.Persistent)   // MOTOR: Speed 10
		c.cvStore.SetDefault(77, 0, store.Persistent)   // MOTOR: Speed 11
		c.cvStore.SetDefault(78, 0, store.Persistent)   // MOTOR: Speed 12
		c.cvStore.SetDefault(79, 0, store.Persistent)   // MOTOR: Speed 13
		c.cvStore.SetDefault(80, 0, store.Persistent)   // MOTOR: Speed 14
		c.cvStore.SetDefault(81, 0, store.Persistent)   // MOTOR: Speed 15
		c.cvStore.SetDefault(82, 0, store.Persistent)   // MOTOR: Speed 16
		c.cvStore.SetDefault(83, 0, store.Persistent)   // MOTOR: Speed 17
		c.cvStore.SetDefault(84, 0, store.Persistent)   // MOTOR: Speed 18
		c.cvStore.SetDefault(85, 0, store.Persistent)   // MOTOR: Speed 19
		c.cvStore.SetDefault(86, 0, store.Persistent)   // MOTOR: Speed 20
		c.cvStore.SetDefault(87, 0, store.Persistent)   // MOTOR: Speed 21
		c.cvStore.SetDefault(88, 0, store.Persistent)   // MOTOR: Speed 22
		c.cvStore.SetDefault(89, 0, store.Persistent)   // MOTOR: Speed 23
		c.cvStore.SetDefault(90, 0, store.Persistent)   // MOTOR: Speed 24
		c.cvStore.SetDefault(91, 0, store.Persistent)   // MOTOR: Speed 25
		c.cvStore.SetDefault(92, 0, store.Persistent)   // MOTOR: Speed 26
		c.cvStore.SetDefault(93, 0, store.Persistent)   // MOTOR: Speed 27
		c.cvStore.SetDefault(94, 0, store.Persistent)   // MOTOR: Speed 28
		c.cvStore.SetDefault(95, 128, store.Persistent) // MOTOR: Reverse trim

		c.cvStore.SetDefault(105, 0, store.Persistent) // MISC: User identification number
		c.cvStore.SetDefault(106, 0, store.Persistent) // MISC: User identification number

		c.cvStore.SetDefault(109, fwVersion[0], roPersist) // SYS: Major version number
		c.cvStore.SetDefault(110, fwVersion[1], roPersist) // SYS: Minor version number
		c.cvStore.SetDefault(111, fwVersion[2], roPersist) // SYS: Patch version number

		c.cvStore.SetDefault(113, 32, store.Persistent) // SYS: Watchdog timeout in 32.768ms steps (1-255)

		c.cvStore.SetDefault(116, 50, store.Persistent)  // MOTOR: Speed step 1 back EMF measurement interval in 0.1ms steps (50-200)
		c.cvStore.SetDefault(117, 150, store.Persistent) // MOTOR: Speed step max back EMF measurement interval in 0.1ms steps (50-200)
		c.cvStore.SetDefault(118, 15, store.Persistent)  // MOTOR: Speed step 1 back EMF measurement cutout duration in 0.1ms steps (10-40)
		c.cvStore.SetDefault(119, 20, store.Persistent)  // MOTOR: Speed step max back EMF measurement cutout duration in 0.1ms steps (10-40)
		// case 1:
		// CVs 257-512
	}

	_, err := c.cvStore.LoadIndex(c.IndexPage(cv31, cv32))
	if err != nil {
		return err
	}

	return nil
}
