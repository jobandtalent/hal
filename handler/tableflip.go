package handler

import (
	"github.com/jobandtalent/hal"
)

// TableFlip is an example of a Handler
var TableFlip = &hal.Handler{
	Method:  hal.HEAR,
	Pattern: `tableflip`,
	Run: func(res *hal.Response) error {
		return res.Send(`(╯°□°）╯︵ ┻━┻`)
	},
}
