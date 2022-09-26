package middleware

import (
	"easyApp/config"
	"easyApp/core"
	"golang.org/x/time/rate"
	"sync"
	"time"
)

type RequestLimitService struct {
	Interval time.Duration
	MaxCount int
	Lock     sync.Mutex
	ReqCount int
}

func newRequestLimitService(interval time.Duration, maxCnt int) *RequestLimitService {
	reqLimit := &RequestLimitService{
		Interval: interval,
		MaxCount: maxCnt,
	}

	go func() {
		ticker := time.NewTicker(interval)
		for {
			<-ticker.C
			reqLimit.Lock.Lock()
			reqLimit.ReqCount = 0
			reqLimit.Lock.Unlock()
		}
	}()

	return reqLimit
}

func (reqLimit *RequestLimitService) increase() {
	reqLimit.Lock.Lock()
	defer reqLimit.Lock.Unlock()

	reqLimit.ReqCount += 1
}

func (reqLimit *RequestLimitService) isAvailable() bool {
	reqLimit.Lock.Lock()
	defer reqLimit.Lock.Unlock()

	return reqLimit.ReqCount < reqLimit.MaxCount
}

var requestLimitsV2 = newRequestLimitService(3*time.Second, 1)

// RequestLimitV2 限流器1(自己实现的)
func RequestLimitV2(ctx *core.Context) {
	if requestLimitsV2.isAvailable() {
		requestLimitsV2.increase()
	} else {
		ctx.Json(1, "前方拥挤请稍后再试", nil)
		return
	}
	ctx.Next()
}

var requestLimits = rate.NewLimiter(100, config.App.ReqBurst) //限流器2用，每秒生成令牌数，最大令牌数
// RequestLimit 限流器2(官方出的)
func RequestLimit(ctx *core.Context) {
	if !requestLimits.Allow() {
		ctx.Json(1, "前方拥挤请稍后再试", nil)
		return
	}
	ctx.Next()
}
