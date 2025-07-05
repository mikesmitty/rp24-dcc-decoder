package dcc

import (
	"runtime"
	"time"
)

func (d *Decoder) Monitor() {
	// The earlier rev rp2350 decoder's smoothing cap adds ~6us to one half-wave
	// TODO: Make this configurable
	delay := uint32(6)

	tr1Min := tr1MinTime + delay
	tr1Max := tr1MaxTime + delay
	tr0Min := tr0MinTime + delay
	tr0Max := tr0MaxTime + delay

	var bit int
	var out byte

	msg := NewMessage(d.cv, d)

	i := 0
	state := Preamble
	for {
		// Save the wave time readings in the ring buffer so we don't lose any
		for !d.sm.IsRxFIFOEmpty() && !d.buf.Full() {
			d.buf.Put(d.sm.RxGet())
		}
		for d.buf.Used() > 0 {
			// Two cycles per loop / 2 MHz clock = 1 microsecond per tick
			ticks, ok := d.buf.Get()
			if !ok {
				// No data available, sleep and wait for more data to be processed
				break
			}

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

					if msg.XOR() {
						msg.Process()
					} else {
						d.checksumErrorCount++
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
		time.Sleep(100 * time.Microsecond) // Sleep a bit to avoid busy-waiting
	}
}
