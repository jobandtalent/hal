package haltest

import (
	"github.com/jobandtalent/hal"
)

type Robot struct {
	hal.Robot
}

type ResponseRecorder struct {
}

func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{}
}
