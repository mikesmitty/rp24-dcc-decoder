package hal

type SimplePWM struct {
	channel uint8
	pwm     PWM
	slice   uint8
	top     float32
}
