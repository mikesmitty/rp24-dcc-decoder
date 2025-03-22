package hal

/* TODO: Implement this
func (h *HAL) SetCapCharging(enable bool) {
	// If we're turning off the capacitor charge control pin, wait until we've received a move request again
	if !enable {
		h.SetPWM("capCharge", 0)
		h.capChargeReady = false
	}

	// No capacitor charge control pin, no charge control
	if _, ok := h.Pins["capCharge"]; !ok {
		return
	}

	// Wait until we're clear to charge (i.e. we've received a move command and aren't on programming track)
	if !h.capChargeReady {
		return
	}

	if enable && h.capChargeReady {
		h.SetPWM("capCharge", CapChargeDuty)
	}
}
*/
