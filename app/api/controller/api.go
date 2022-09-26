package controller

import (
	"easyApp/config"
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
	var aaa Ass
	_, err := ctx.AuthBodyBindJSON(&aaa, true)
	if err != nil {
		ctx.Json(1, err, aaa)
		return
	}

	ctx.Json(0, "ok", config.CallHost)
	return
}
