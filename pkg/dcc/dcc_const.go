package dcc

import (
	"time"

	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/motor"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/ringbuffer"
)

const (
	// 2 MHz PIO clock for easy timing calculations
	// 2 instructions per counter increment gives the counter in microseconds
	smFreq = 2_000_000

	maxMsgLength     = 11
	preambleLength   = 11
	rcCutoutStartMin = 26
	rcCutoutStartMax = 32
	tr1MinTime       = 52
	tr1MaxTime       = 64
	tr0MinTime       = 90
	tr0MaxTime       = 10_000
)

type Decoder struct {
	cv    cv.Handler
	motor *motor.Motor

	sm     shared.StateMachine
	offset uint8
	buf    *ringbuffer.RingBuffer[uint32]

	address        []byte
	consistAddress []byte
	Snoop          bool

	capPin     shared.Pin
	outputPins []shared.Pin
	rcTxPin    shared.Pin
	rcTxQueued bool

	opMode           opMode
	lastSvcResetTime time.Time
	svcModeReady     bool

	outputCallbacks map[uint16][]shared.OutputCallback
	outputMapsFwd   map[uint16]uint16
	outputMapsRev   map[uint16]uint16

	consistFuncMask [3]uint8

	lastDirection motor.Direction
}

type decoderState int

const (
	Preamble decoderState = iota
	Bits
	Terminator
)

type opMode int

const (
	UndefinedMode opMode = iota
	OperationsMode
	ServiceMode
)
