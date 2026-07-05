package motor

const (
	// BEMF sense voltage divider: motor terminal -> 6.8k -> ADC pin -> 1k -> GND
	// (R16/R17 and R18/R19 on the rp2350-decoder board)
	bemfDividerRatio = 1.0 / (6.8 + 1.0) // ~0.128, so 3.3V ADC full scale = ~25.7V at the motor
	adcRefVolts      = 3.3

	// ADC counts per volt of motor-terminal BEMF (~2546)
	bemfCountsPerVolt = bemfDividerRatio * 65535 / adcRefVolts
)

// SpeedMode defines the type of speed step mode
type SpeedMode uint8

const (
	// The value of each SpeedMode is the number of speed steps excluding stop and e-stop
	SpeedMode14  SpeedMode = 14
	SpeedMode28  SpeedMode = 28
	SpeedMode128 SpeedMode = 126
)

type Direction uint8

const (
	Reverse Direction = iota
	Forward
)

func (d Direction) String() string {
	switch d {
	case Reverse:
		return "Reverse"
	case Forward:
		return "Forward"
	default:
		return "Unknown"
	}
}
