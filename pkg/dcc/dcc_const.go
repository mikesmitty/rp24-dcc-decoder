package dcc

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
