package middleware

import (
	"easyApp/core"
	"net/http"
)

// Cors 解决跨域
func Cors(ctx *core.Context) {
	// 必须，接受指定域的请求，可以使用*不加以限制，但不安全
	ctx.Header("Access-Control-Allow-Origin", ctx.GetHeader("Origin"))
	// 必须，设置服务器支持的所有跨域请求的方法
	ctx.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	// 服务器支持的所有头信息字段，不限于浏览器在"预检"中请求的字段
	ctx.Header("Access-Control-Allow-Headers", "Content-Type,X-Requested-With ,Content-Length")
	// 可选，设置XMLHttpRequest的响应对象能拿到的额外字段
	ctx.Header("Access-Control-Expose-Headers", "Content-Type, Access-Control-Allow-Headers")
	// 可选，是否允许后续请求携带认证信息cookie，该值只能是true，不需要则不设置
	//ctx.Header("Access-Control-Allow-Credentials", "true")

	// 放行所有OPTIONS方法
	if ctx.Request.Method == "OPTIONS" {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}
	ctx.Next()
}
