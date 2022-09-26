package middleware

import (
	"easyApp/config"
	"easyApp/core"
	"errors"
	"fmt"
	"runtime/debug"
)

// RecoveryJSON 恐慌捕获，并已JSON格式响应
func RecoveryJSON(ctx *core.Context) {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case error:
				err = err.(error).Error()
				break
			case []byte:
				err = string(err.([]byte))
			default:
				err = fmt.Sprint(err)
				break
			}

			if config.AppMode() != "release" {
				ctx.Json(50000, err.(string)+" \n"+string(debug.Stack()), nil)
			} else {
				ctx.Json(50000, "哎呀出错了~", nil)
			}
			ctx.ErrPush(errors.New(err.(string) + " \n" + string(debug.Stack())))
		}
	}()
	ctx.Next()
}
