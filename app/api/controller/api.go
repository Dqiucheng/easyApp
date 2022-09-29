package controller

import (
	"easyApp/core"
)

type Api struct {
}

type Ass struct {
	Aa   string `validate:"required"`
	Ut   string `validate:"required"`
	Ut_a string `validate:"required"`
}

func (Api) Test(ctx *core.Context) {

	ctx.Json(0, "ok", nil)
	return
}
