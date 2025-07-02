//go:build !rp

package shared

// Avoid requiring packages that require specific hardware so we can run unit tests

const (
	KHz = 1_000
	MHz = 1_000_000

	NoPin = MockPin(0xff)
)

// package rp2-pio
type I2S interface {
	SetSampleFrequency(f uint32) error
	WriteMono(data []uint16) (int, error)
}

// package machine
type Pin interface {
	Configure(MockPinConfig)
	Get() bool
	High()
	Low()
	Set(bool)
	SetInterrupt(MockPinChange, func(Pin)) error
}

// package machine
type PinConfig any

// package machine
type PWMConfig any

// package rp2-pio
type StateMachine interface {
	IsRxFIFOEmpty() bool
	IsRxFIFOFull() bool
	RxFIFOLevel() uint32
	RxGet() uint32
	SetEnabled(bool)
}

type CVCallbackFunc func(uint16, uint8) bool

type OutputCallback func(uint16, bool)

type MockPin uint8

func (m MockPin) Configure(mode MockPinConfig) {}

func (m MockPin) Get() bool {
	return false
}

func (m MockPin) High() {}

func (m MockPin) Low() {}

func (m MockPin) Set(bool) {}

func (m MockPin) SetInterrupt(MockPinChange, func(Pin)) error {
	return nil
}

type MockPinChange uint8

type MockPinConfig struct {
	Mode MockPinMode
}

type MockPinMode uint8
