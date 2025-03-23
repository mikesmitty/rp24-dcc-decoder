//go:build !rp

package motor

import (
	"github.com/mikesmitty/rp24-dcc-decoder/internal/shared"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/cv"
	"github.com/mikesmitty/rp24-dcc-decoder/pkg/hal"
)

type Motor struct {
	ndotReverse bool
	reverse     bool
}

func NewMotor(conf cv.Handler, hw *hal.HAL, pinA, pinB, emfA, emfB, adcRef shared.Pin) *Motor {
	return &Motor{}
}

func (m *Motor) SetSpeed(speed uint8, reverse bool) {
}
