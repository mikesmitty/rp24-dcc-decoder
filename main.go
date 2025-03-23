package main

import (
	"runtime"

	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/dcc"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/hal"
)

var hw *hal.HAL

func main() {
	hw = hal.NewHAL()
	cvHandler := cv.NewCVHandler([]uint8{1, 2, 3}) // TODO: Set version number at build time

	dccPin, ok := hw.Pin("dcc")
	if !ok {
		panic("DCC pin not found")
	}

	println("Starting DCC") // FIXME: Cleanup?
	pioNum := 0
	d, err := dcc.NewDecoder(cvHandler, pioNum, dccPin)
	if err != nil {
		panic(err.Error())
	}

	// Register available outputs
	for _, output := range outputs {
		if _, ok := hw.Pin(output); ok {
			d.RegisterOutput(output, hw.GetOutputCallback(output))
		}
	}

	// FIXME: Initialize motor controller

	d.SetAddress(150) // FIXME: Cleanup
	// FIXME: Avoid blocking main thread
	d.Monitor()

	for {
		runtime.Gosched() // FIXME: Cleanup
	}
}

var outputs = []string{
	"lampFront",
	"lampRear",
	"aux1",
	"aux2",
	"aux3",
	"aux4",
	"aux5",
	"aux6",
	"aux7",
	"aux8",
	"aux10",
	"aux11",
	"aux12",
}
