package hal

import "github.com/mikesmitty/rp24-dcc-decoder/internal/shared"

func (h *HAL) GetOutputCallback(output string) shared.OutputCallback {
	return func(_ uint16, on bool) {
		h.SetOutput(output, on)
	}
}

func (h *HAL) SetOutput(output string, state bool) {
	// TODO: Add support for dimming, strobing, etc.
	if pin, ok := h.pins[output]; ok {
		pin.Set(state)
	}
}
