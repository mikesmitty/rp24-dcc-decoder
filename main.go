package main

import (
	"strconv"
	"strings"

	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/dcc"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/hal"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/motor"
)

var hw *hal.HAL

var version = "1.1.1"

func main() {
	hw = hal.NewHAL()
	cvHandler := cv.NewCVHandler(versionToBytes(version))

	dccPin, ok := hw.PinOk("dcc")
	if !ok {
		panic("DCC pin not found")
	}

	motorA, okA := hw.PinOk("motorA")
	motorB, okB := hw.PinOk("motorB")
	emfA, okEA := hw.PinOk("emfA")
	emfB, okEB := hw.PinOk("emfB")
	adcRef, okADC := hw.PinOk("adcRef")
	if !okA || !okB || !okEA || !okEB || !okADC {
		panic("Motor pins not found")
	}

	m := motor.NewMotor(cvHandler, hw, motorA, motorB, emfA, emfB, adcRef)

	outputPins := make([]shared.Pin, 0, len(outputs))
	for _, output := range outputs {
		if pin, ok := hw.PinOk(output); ok {
			outputPins = append(outputPins, pin)
		}
	}

	println("Starting DCC")
	capPin := hw.Pin("capCharge")
	rcTxPin := hw.Pin("railcom")
	pioNum := 0
	d, err := dcc.NewDecoder(cvHandler, m, pioNum, dccPin, capPin, rcTxPin, outputPins)
	if err != nil {
		panic(err.Error())
	}

	// Register available outputs
	for _, output := range outputs {
		if _, ok := hw.PinOk(output); ok {
			d.RegisterOutput(output, hw.GetOutputCallback(output))
		}
	}

	go m.Run()
	go m.RunEMF()
	d.Monitor()
}

func versionToBytes(version string) []uint8 {
	versionParts := strings.Split(version, ".")
	if len(versionParts) != 3 {
		panic("invalid version string length")
	}
	var versionBytes []uint8
	for _, part := range versionParts {
		partInt, err := strconv.Atoi(part)
		if err != nil {
			panic("invalid version string")
		}
		versionBytes = append(versionBytes, uint8(partInt))
	}
	return versionBytes
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
