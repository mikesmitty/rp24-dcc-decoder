//go:build rp

package hal

import (
	"errors"
	"machine"

	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
	pio "github.com/tinygo-org/pio/rp2-pio"
	"github.com/tinygo-org/pio/rp2-pio/piolib"
)

func (h *HAL) initI2SPIO(pioNum int, dataPin, BCLKPin machine.Pin) (shared.I2S, error) {
	var sm pio.StateMachine
	var err error

	switch pioNum {
	case 0:
		sm, err = pio.PIO0.ClaimStateMachine()
	case 1:
		sm, err = pio.PIO1.ClaimStateMachine()
	case 2:
		// sm, err = pio.PIO2.ClaimStateMachine()
		return nil, errors.New("PIO2 not yet supported")
	}
	if err != nil {
		return nil, err
	}
	// Pio := sm.PIO()

	i2s, err := piolib.NewI2S(sm, dataPin, BCLKPin)
	if err != nil {
		panic(err.Error())
	}

	return i2s, nil
}
