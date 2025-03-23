package dcc

import (
	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/motor"
)

func (d *Decoder) RegisterOutput(output string, fn shared.OutputCallback) {
	index := IndexFromOutput(output)
	d.outputCallbacks[index] = append(d.outputCallbacks[index], fn)
}

// Control DCC functions
func (d *Decoder) callFunction(number uint16, on bool) {
	var ok bool
	// If there's no separate reverse callback use forward instead
	outputMap, hasReverse := d.outputMapsRev[number]
	direction := d.motor.Direction()
	if direction == motor.Forward || !hasReverse {
		outputMap, ok = d.outputMapsFwd[number]
	}
	if !ok {
		println("output map not found:", number)
		return
	}

	// If the direction has changed from reverse to forward and this function has separate forward/reverse
	// output maps turn off the outputs that would have been on in reverse. Not needed if there's no
	// separate reverse output map, as the forward map is used for both directions
	if d.lastDirection != direction && hasReverse {
		d.lastDirection = direction

		outputMapPrev := d.outputMapsRev[number]
		if direction == motor.Reverse {
			outputMapPrev = d.outputMapsFwd[number]
		}
		for i := 0; i < 16; i++ {
			if outputMapPrev&(1<<i) != 0 {
				// Turn off all the "on" outputs from the previous direction
				if handlers, ok := d.outputCallbacks[number]; ok {
					for _, fn := range handlers {
						fn(outputMap, false)
					}
				}
			}
		}
	}

	for i := 0; i < 16; i++ {
		if outputMap&(1<<i) != 0 {
			if handlers, ok := d.outputCallbacks[number]; ok {
				for _, fn := range handlers {
					fn(number, on)
				}
			}
		}
	}
}

func IndexFromOutput(output string) uint16 {
	switch output {
	case "lampFront":
		return 0
	case "lampRear":
		return 1
	case "aux1":
		return 2
	case "aux2":
		return 3
	case "aux3":
		return 4
	case "aux4":
		return 5
	case "aux5":
		return 6
	case "aux6":
		return 7
	case "aux7":
		return 8
	case "aux8":
		return 9
	case "aux10":
		return 11
	case "aux11":
		return 12
	case "aux12":
		return 13
	}
	return 255
}
