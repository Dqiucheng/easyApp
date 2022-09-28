package logger

import (
	"context"
	"easyApp/config"
	"fmt"
	"github.com/Dqiucheng/dlogroller"
	"github.com/gin-gonic/gin"
	json "github.com/goccy/go-json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"path"
)

var (
	logger    *zap.Logger // 主要用来记录系统 之外 产生的相关日志
	sysLogger *zap.Logger // 主要用来记录系统产生的相关日志
)

var (
	encoderConfig zapcore.Encoder
	sync          []zapcore.WriteSyncer // 输入源
	atomicLevel   = zap.NewAtomicLevel()
)

func init() {
	// 是否同时输出到控制台，1是、0否
	if config.App.DLogIsInStdout == 1 || config.AppMode() == "debug" {
		sync = append(sync, zapcore.AddSync(os.Stdout))
	}

	// 设置日志级别
	switch config.App.DLogSetLevel {
	case "panic":
		atomicLevel.SetLevel(zap.PanicLevel)
		break
	case "fatal":
		atomicLevel.SetLevel(zap.FatalLevel)
		break
	case "error":
		atomicLevel.SetLevel(zap.ErrorLevel)
		break
	case "warn", "warning":
		atomicLevel.SetLevel(zap.WarnLevel)
		break
	case "info":
		atomicLevel.SetLevel(zap.InfoLevel)
		break
	case "debug":
		atomicLevel.SetLevel(zap.DebugLevel)
		break
	default:
		atomicLevel.SetLevel(zap.InfoLevel)
		break
	}

	encoderConfig = zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		LevelKey:       "lv",
		TimeKey:        "ts",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     "\n",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,                          // 小写编码器
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"), // TimeKey的时间格式，zapcore.ISO8601TimeEncoder UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,                         // 一般zapcore.SecondsDurationEncoder,执行消耗的时间转化成浮点型的秒
		EncodeCaller:   zapcore.FullCallerEncoder,                              // 全路径编码器，一般zapcore.ShortCallerEncoder以包/文件:行号 格式化调用堆栈
		EncodeName:     zapcore.FullNameEncoder,
	})

	// 创建*zap.Logger
	NewLogger(&logger, "")
	NewLogger(&sysLogger, "sys")
}

// NewLogger 创建一个*zap.Logger
// 可利用该函数在任何地方自定义一个Logger，主要用来解决日志的分类，比如创建一个SqlLogger
// 避免重复创建同一个*zap.Logger重复创建，在降低性能的同时还可能会造成文件的竞争
// 可重复使用创建的loggers，但不可重复创建
//
// loggers 要创建的*zap.Logger变量
// module 日志模块（目录），可为空。DLogRootPath/appName/module/200601/02/15.log
func NewLogger(loggers **zap.Logger, module string) {
	// 日志级别分割
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		if lvl > zapcore.InfoLevel {
			return false
		}
		return lvl >= atomicLevel.Level()
	})
	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		if lvl <= zapcore.InfoLevel {
			return false
		}
		return lvl >= atomicLevel.Level()
	})

	core := zapcore.NewTee(
		zapcore.NewCore(
			encoderConfig, // 编码器配置
			zapcore.NewMultiWriteSyncer(append(sync, zapcore.AddSync(newDlogRoller(module, "")))...), // 输入方式
			infoLevel, // 日志级别
		),
		zapcore.NewCore(
			encoderConfig, // 编码器配置
			zapcore.NewMultiWriteSyncer(append(sync, zapcore.AddSync(newDlogRoller(module, "_err")))...), // 输入方式
			errorLevel, // 日志级别
		),
	)

	*loggers = zap.New(
		core,
		zap.AddCaller(),
	)
	return
}

// newDlogRoller 日志切割
func newDlogRoller(module, Level string) *dlogroller.Roller {
	// 日志跟目录
	DLogRootPath := config.App.DLogRootPath
	if DLogRootPath == "" {
		getwdPath, _ := os.Getwd()
		DLogRootPath = getwdPath + "/logs/"
	}

	hook, err := dlogroller.New(
		path.Join(DLogRootPath, config.App.AppName, module),
		path.Join("%Y%m", "%d", "%m-%dT%H"+Level+".log"),
		dlogroller.SetMaxSize(config.App.DLogMaxSize),
		dlogroller.SetMaxAge(config.App.DLogMaxAge),
		dlogroller.SetMillEveryDayHour(config.App.DLogMillEveryDayHour),
	)
	if err != nil {
		log.Panicf("创建dlogroller失败: %s", err)
		return nil
	}
	return hook
}

// packField 默认日志字段组装
func packField(logger *zap.Logger, logData interface{}) *zap.Logger {
	if logData == nil {
		return logger.With(zap.String("logData", ""))
	}

	switch logData.(type) {
	case string: // 由于json数据会被多次编码因此需要提前处理，提高最终可读性，但是会因此略微降低性能
		if len(logData.(string)) > 0 {
			if logDataT := logData.(string)[0:1]; logDataT == `{` || logDataT == `[` {
				if json.Unmarshal([]byte(logData.(string)), &logData) == nil {
					return logger.With(zap.Reflect("logData", logData))
				}
			}
		}
		return logger.With(zap.String("logData", logData.(string)))
	case []byte:
		return logger.With(zap.String("logData", string(logData.([]byte))))
	}

	return logger.With(zap.Any("logData", logData))
}

/*********************************************************************************************************************/

// L 返回一个*zap.Logger
func L() *zap.Logger {
	return logger
}

// Log 		记录日志
//
// logData	日志数据
// 使用方式	Log("我是想要记录的数据").Info("备注")
func Log(logData interface{}) *zap.Logger {
	return packField(logger, logData)
}

// LogContext 		记录包含请求ID的日志
func LogContext(ctx *gin.Context, logData interface{}) *zap.Logger {
	return packField(logger, logData).With(
		zap.Int64("reqId", ctx.GetInt64("ReqId")),
		zap.String("reqURI", ctx.Request.RequestURI),
	)
}

/*********************************************************************************************************************/

// SysL 返回一个系统*zap.Logger
func SysL() *zap.Logger {
	return sysLogger
}

// SysLog 系统日志，主要用来记录系统产生的相关日志
func SysLog(logData interface{}) *zap.Logger {
	return packField(sysLogger, logData)
}

// SysLogContext 		记录包含请求ID的日志
func SysLogContext(ctx *gin.Context, logData interface{}) *zap.Logger {
	return packField(logger, logData).With(zap.Int64("reqId", ctx.GetInt64("ReqId")))
}

// ErrPush 自定义错误信息通知》》》
func ErrPush(ctx context.Context, errMsg error) {
	if config.AppMode() != "debug" {
		fmt.Println(ctx, errMsg)
	}

	return
}
