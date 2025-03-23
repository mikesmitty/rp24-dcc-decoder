package hal

func (h *HAL) InitI2S() error {
	// LRCLKPin is expected to be BCLKPin + 1
	i2s, err := h.initI2SPIO(0, h.pins["i2sDIN"], h.pins["i2sBLCK"])
	if err != nil {
		return err
	}

	// TODO: Make configurable
	i2s.SetSampleFrequency(44100) // TODO: Use 48kHz or 96kHz for ultrasonic frequencies

	/* TODO: Implement
	i2s.WriteStereo(data)
	}
	*/
	return nil
}
