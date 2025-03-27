package main

import (
	"strconv"
	"strings"

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

	dccPin, ok := hw.Pin("dcc")
	if !ok {
		panic("DCC pin not found")
	}

	motorA, okA := hw.Pin("motorA")
	motorB, okB := hw.Pin("motorB")
	emfA, okEA := hw.Pin("emfA")
	emfB, okEB := hw.Pin("emfB")
	adcRef, okADC := hw.Pin("adcRef")
	if !okA || !okB || !okEA || !okEB || !okADC {
		panic("Motor pins not found")
	}

	m := motor.NewMotor(cvHandler, hw, motorA, motorB, emfA, emfB, adcRef)

	println("Starting DCC")
	pioNum := 0
	d, err := dcc.NewDecoder(cvHandler, m, pioNum, dccPin)
	if err != nil {
		panic(err.Error())
	}

	// Register available outputs
	for _, output := range outputs {
		if _, ok := hw.Pin(output); ok {
			d.RegisterOutput(output, hw.GetOutputCallback(output))
		}
	}

	d.SetAddress(150) // FIXME: Cleanup
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
