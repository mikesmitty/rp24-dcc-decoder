package hal

type PWM interface {
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
	Enable(bool)
	Top() uint32
	Configure(config PWMConfig) error
	Channel(Pin) (uint8, error)
}

type PWMConfig interface{}

type SimplePWM struct {
	channel uint8
	pwm     PWM
	slice   uint8
	top     float32
}
