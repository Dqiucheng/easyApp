package app

import (
	"easyApp/app/api/controller"
	"easyApp/core"
	"easyApp/middleware"
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化应用路由
func InitRouter(router *gin.Engine)  {
	ApiRouter(router)
}

// Router Api路由注册
func ApiRouter(router *gin.Engine) {
	apiGroup := router.Group("/api")
	apiGroup.Use(core.Handle(middleware.RequestLimit))
	{
		api := new(controller.Api)
		apiGroup.Any("/", core.Handle(api.Test))
	}
}
