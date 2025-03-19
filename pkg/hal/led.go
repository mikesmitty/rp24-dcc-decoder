//go:build rp

package hal

func (h *HAL) StatusLED(enable bool) {
	if enable {
		h.pins["led"].High()
	} else {
		h.pins["led"].Low()
	}
}
