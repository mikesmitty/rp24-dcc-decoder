//go:build rp

package dcc

import (
	"errors"
	"machine"

	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
	pio "github.com/tinygo-org/pio/rp2-pio"
)

func (d *Decoder) initPIO(pioNum int, p shared.Pin) error {
	var sm pio.StateMachine
	var err error
	pin := p.(machine.Pin)

	switch pioNum {
	case 0:
		sm, err = pio.PIO0.ClaimStateMachine()
	case 1:
		sm, err = pio.PIO1.ClaimStateMachine()
	case 2:
		// sm, err = pio.PIO2.ClaimStateMachine()
		return errors.New("PIO2 not yet supported")
	}
	if err != nil {
		return err
	}
	Pio := sm.PIO()

	offset, err := Pio.AddProgram(dccInstructions, dccOrigin)
	if err != nil {
		return err
	}

	whole, frac, err := pio.ClkDivFromFrequency(smFreq, machine.CPUFrequency())
	if err != nil {
		return err
	}

	// Enable the internal pull-up resistor first, the circuit will pull down when negative
	pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	// Switching to PIO mode clears most pin settings, but doesn't change the pull-up/pull-down state
	pin.Configure(machine.PinConfig{Mode: Pio.PinMode()})

	cfg := dccProgramDefaultConfig(offset)
	// Disable autopush
	cfg.SetInShift(false, false, 32)
	// Set our GPIO to be used in JMP PIN commands
	cfg.SetJmpPin(pin)
	// Combine the TX/RX FIFO buffers to allow extra breathing room between buffer reads
	cfg.SetFIFOJoin(pio.FifoJoinRx)

	sm.Init(offset, cfg)
	sm.SetClkDiv(whole, frac)

	d.sm = sm
	d.offset = offset
	d.Enable(true)

	return nil
}
