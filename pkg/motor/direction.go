package motor

// Direction returns the current direction of the motor
func (m *Motor) Direction() Direction {
	// Double-reverse or double-forward both equal forward
	if m.reverse == m.ndotReverse {
		return Forward
	}
	return Reverse
}
