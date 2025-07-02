package hal

import "github.com/mikesmitty/rp24-dcc-decoder/internal/shared"

func (h *HAL) PinOk(name string) (shared.Pin, bool) {
	_, ok := h.pins[name]
	return h.pins[name], ok
}

func (h *HAL) Pin(name string) shared.Pin {
	pin, ok := h.pins[name]
	if !ok {
		return shared.NoPin
	}
	return pin
}
