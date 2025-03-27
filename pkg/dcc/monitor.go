package dcc

import "runtime"

func (d *Decoder) Monitor() {
	var bit int
	var out byte
	var ticks uint32

	msg := NewMessage(d.cv, d)

	i := 0
	state := Preamble
	for {
		if !d.sm.IsRxFIFOEmpty() {
			// Two cycles per loop / 2 MHz clock = 1 microsecond per tick
			ticks = d.sm.RxGet()

			// Make sure the bit is within the valid ranges
			bit = 0
			if ticks >= tr1Min && ticks <= tr1Max {
				bit = 1
			} else if ticks < tr0Min || ticks > tr0Max {
				// Noise
				continue
			}

			// Increment the bit counter
			i++

			switch state {
			case Preamble:
				// Waiting for the preamble count to be satisfied
				if bit == 0 {
					if i >= preambleLength {
						// Preamble terminator bit received, ready to decode data
						state = Bits
						i = 0
					} else {
						// Reset the bit counter, we got a premature terminating bit
						i = 0
					}
				}
			case Bits:
				// Shift in a zero bit
				out <<= 1
				// Set it to 1
				if bit == 1 {
					out |= 1
				}
				// Next bit will be a terminator, save byte and prep for the next loop
				if i == 8 {
					msg.AddByte(out)
					out = 0
					state = Terminator
				}
			case Terminator:
				i = 0
				state = Bits

				if bit == 1 {
					// End of message terminator bit received, wait for a preamble next loop
					state = Preamble

					if !msg.XOR() {
						for _, b := range msg.Bytes() {
							printByte(b)
							print(" ")
						}
						println("checksum error")
					} else {
						msg.Process()
					}

					// Reset the message buffer for the next message
					msg.Reset()
					runtime.Gosched()
				}
			default:
				println("ERROR: invalid decoder state: ", state)
				state = Preamble
				msg.Reset()
			}
		}
	}
}

// Print a byte as bits
func printByte(b byte) {
	for i := 7; i >= 0; i-- {
		if b&(1<<i) != 0 {
			print("1")
		} else {
			print("0")
		}
	}
}

// Print the n leftmost bits of b as a binary number
func printUintN(n int, b uint32) {
	for i := n - 1; i >= 0; i-- {
		if b&(1<<i) != 0 {
			print("1")
		} else {
			print("0")
		}
	}
}
