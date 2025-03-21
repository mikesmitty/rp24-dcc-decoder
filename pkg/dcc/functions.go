package dcc

type FnHandler func(bool)

// Register a function handler
func (d *Decoder) RegisterFunction(number uint16, fn FnHandler) {
	d.fnHandlers[number] = append(d.fnHandlers[number], fn)
}

// Control DCC functions
func (d *Decoder) callFunction(number uint16, state bool) {
	if handlers, ok := d.fnHandlers[number]; ok {
		for _, fn := range handlers {
			fn(state)
		}
	} else {
		println("function not found:", number)
	}
}
