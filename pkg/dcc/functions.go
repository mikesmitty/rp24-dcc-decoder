package dcc

import (
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cb"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/motor"
)

func (d *Decoder) RegisterOutput(output string, fn cb.OutputCallback) {
	index := IndexFromOutput(output)
	d.outputCallbacks[index] = append(d.outputCallbacks[index], fn)
}

// Control DCC functions
func (d *Decoder) callFunction(number uint16, state bool) {
	var outputMap uint16
	var ok bool
	if d.direction == motor.Forward {
		outputMap, ok = d.outputMapsFwd[number]
	} else {
		outputMap, ok = d.outputMapsRev[number]
	}
	if !ok {
		println("output map not found:", number)
		return
	}

	for i := 0; i < 16; i++ {
		if outputMap&(1<<i) > 0 {
			if handlers, ok := d.outputCallbacks[number]; ok {
				for _, fn := range handlers {
					fn(number, state)
				}
			} else {
				println("function not found:", number)
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
