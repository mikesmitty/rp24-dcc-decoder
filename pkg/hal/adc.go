//go:build rp

package hal

import (
	"machine"

	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
)

type ADC struct {
	adc machine.ADC
}

func NewADC(pin shared.Pin) *ADC {
	machine.InitADC()

	a := machine.ADC{Pin: pin.(machine.Pin)}
	a.Configure(machine.ADCConfig{})

	return &ADC{adc: a}
}

func (a *ADC) Read() uint16 {
	return a.adc.Get()
}
