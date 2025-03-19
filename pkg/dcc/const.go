package dcc

import (
	"time"

	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/motor"
)

const (
	// 2 MHz PIO clock for easy timing calculations
	// 2 instructions per counter increment gives the counter in microseconds
	smFreq = 2_000_000

	preambleLength = 11
	tr1Min         = 52
	tr1Max         = 64
	tr0Min         = 90
	tr0Max         = 10_000
)

type Decoder struct {
	cv cv.Handler

	sm     StateMachine
	offset uint8

	address        []byte
	consistAddress []byte
	Snoop          bool

	opMode           opMode
	lastSvcResetTime time.Time
	svcModeReady     bool

	fnHandlers map[uint16][]FnHandler

	// FIXME: Add CV callback to adjust?
	// FIXME: Move to cv package?
	speedMode motor.SpeedMode
}

type Pin interface{}

// Avoid requiring the rp2-pio package so we can run unit tests
type StateMachine interface {
	IsRxFIFOEmpty() bool
	RxGet() uint32
	SetEnabled(bool)
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
