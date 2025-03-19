//go:build rp

package hal

import "machine"

type ADC struct {
	adc machine.ADC
}

func NewADC(pin machine.Pin) *ADC {
	machine.InitADC()

	a := machine.ADC{Pin: pin}
	a.Configure(machine.ADCConfig{})

	return &ADC{
		adc: a,
	}
}

func (a *ADC) Read() uint16 {
	return a.adc.Get()
}
