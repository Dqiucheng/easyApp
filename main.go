package main

import (
	"easyApp/config"
	"easyApp/db"
	"easyApp/logger"
	"easyApp/router"
	"easyApp/util"
	"context"
	"errors"
	"github.com/DeanThompson/ginpprof"
	"github.com/Dqiucheng/httpClient"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	appMode := config.AppMode()
	if appMode == "" {
		logger.SysLog(nil).Panic("Server日志：app配置文件 SetMode 必须设置")
	}
	// 设置模式
	gin.SetMode(appMode)

	// 禁用控制台中的颜色输出
	gin.DisableConsoleColor()

	// 初始化httpClient客户端请求超时时间，后期httpClient的请求相关操作都基于此参数。
	httpClient.SetTimeout(time.Duration(20))

	// 创建无任何中间件的核心驱动
	ginEngine := gin.New()

	// 非正式环境注册项
	if appMode != "release" {
		// 启动性能监控
		ginpprof.Wrap(ginEngine)
	}

	// 注册db
	createDB()

	// 注册路由（自定义中间件在这里设置）
	router.RunRouter(ginEngine)

	// 启动服务
	httpServerRun(ginEngine)
}

// httpServerRun 启动服务
func httpServerRun(router *gin.Engine) {
	httpPort := config.App.HttpPort
	if httpPort == "" {
		logger.SysLog(nil).Panic("Server日志：app配置文件 HttpPort 必须设置")
	}

	isHttp := false                 // 是否为HTTPS服务
	certFile := config.App.CertFile // .crt文件路径
	keyFile := config.App.KeyFile   // .key文件路径
	if certFile != "" && keyFile != "" {
		if util.Exists(certFile) {
			isHttp = true
		} else {
			logger.SysLog(nil).Panic("Server日志：HTTPS证书路径有误")
		}
	}

	srv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: router,
	}

	// 在协成中初始化服务，这样它就不会堵塞下面优雅的关闭处理
	go func() {
		// 服务连接
		if isHttp {
			logger.SysLog("服务启动成功HTTPS on " + httpPort).Info("Server日志")
			if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.SysLog("listen: " + err.Error()).Fatal("Server日志")
			}
		} else {
			logger.SysLog("服务启动成功HTTP on " + httpPort).Info("Server日志")
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.SysLog("listen: " + err.Error()).Fatal("Server日志")
			}
		}
	}()

	time.Sleep(500 * time.Millisecond)
	_ = logger.L().Sync()
	_ = logger.SysL().Sync()
	log.Println("服务启动 on ", httpPort)

	// 等待中断信号以优雅地关闭服务器（设置 10 秒的超时时间）
	quit := make(chan os.Signal)
	// kill (无参数) 默认发送 syscall.SIGTERM
	// kill -2 是 syscall.SIGINT
	// kill -9 是 syscall.SIGKILL 但是不能被捕获，所以不需要添加它
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 监听关闭信号，并将其写入通道
	<-quit
	logger.SysLog("Shutting down server...").Info("Server日志")

	// 上下文用来通知服务器它有10秒钟的时间来完成当前正在处理的请求
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.SysLog("Server forced to shutdown: " + err.Error()).Fatal("Server日志")
	}

	logger.SysLog("Server exiting").Info("Server日志")
	_ = logger.L().Sync()
	_ = logger.SysL().Sync()
}

// createDB 注册db
func createDB() {
	db.ConnectMySQLS()
	db.ConnectRedisS()
}
