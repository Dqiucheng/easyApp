package middleware

import (
	"easyApp/config"
	"easyApp/core"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
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
				ctx.Json(50000, err.(string)+takeStacktrace(2), nil)
			} else {
				ctx.Json(50000, "哎呀出错了~", nil)
			}
			ctx.ErrPush(errors.New(err.(string) + takeStacktrace(2)))
		}
	}()
	ctx.Next()
}



type programCounters struct {
	pcs []uintptr
}

func newProgramCounters(size int) *programCounters {
	return &programCounters{make([]uintptr, size)}
}


var (
	_stacktracePool = sync.Pool{
		New: func() interface{} {
			return newProgramCounters(64)
		},
	}
)

// takeStacktrace 跳过指定堆栈层级
func takeStacktrace(skip int) string {
	buffer := strings.Builder{}
	programCounters := _stacktracePool.Get().(*programCounters)
	defer _stacktracePool.Put(programCounters)

	var numFrames int
	for {
		// Skip the call to runtime.Callers and takeStacktrace so that the
		// program counters start at the caller of takeStacktrace.
		numFrames = runtime.Callers(skip+2, programCounters.pcs)
		if numFrames < len(programCounters.pcs) {
			break
		}
		// Don't put the too-short counter slice back into the pool; this lets
		// the pool adjust if we consistently take deep stacktraces.
		programCounters = newProgramCounters(len(programCounters.pcs) * 2)
	}

	frames := runtime.CallersFrames(programCounters.pcs[:numFrames])

	// Note: On the last iteration, frames.Next() returns false, with a valid
	// frame, but we ignore this frame. The last frame is a runtime frame which
	// adds noise, since it's only either runtime.main or runtime.goexit.
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		buffer.WriteByte('\n')
		buffer.WriteString(frame.Function)
		buffer.WriteByte('\n')
		buffer.WriteByte('\t')
		buffer.WriteString(frame.File)
		buffer.WriteByte(':')
		buffer.WriteString(strconv.FormatInt(int64(frame.Line), 10))
	}

	return buffer.String()
}
