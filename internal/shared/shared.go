package shared

// Avoid requiring packages that require specific hardware so we can run unit tests

// package rp2-pio
type I2S interface {
	SetSampleFrequency(f uint32) error
	WriteMono(data []uint16) (int, error)
}

// package machine
type Pin interface {
	// Configure(PinConfig)
	Set(bool)
}

// package machine
type PinConfig interface{}

// package machine
type PWMConfig interface{}

// package rp2-pio
type StateMachine interface {
	IsRxFIFOEmpty() bool
	RxGet() uint32
	SetEnabled(bool)
}

type CVCallbackFunc func(uint16, uint8) bool

type OutputCallback func(uint16, bool)
