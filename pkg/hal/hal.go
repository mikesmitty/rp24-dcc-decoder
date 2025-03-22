package hal

func (h *HAL) Pin(name string) (Pin, bool) {
	_, ok := h.pins[name]
	return h.pins[name], ok
}
