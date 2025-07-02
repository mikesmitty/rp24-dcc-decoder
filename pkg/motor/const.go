package motor

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
