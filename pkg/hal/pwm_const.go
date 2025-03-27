package hal

type SimplePWM struct {
	channel uint8
	pwm     pwm
	slice   uint8
	top     float32
}
