package core

import (
	"github.com/gin-gonic/gin"
	"time"
)

type HandlerFunc func(c *Context)

func Handle(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Get("ReqId"); !ok {
			t := time.Now()
			c.Set("StartTime", t)
			c.Set("ReqId", t.UnixNano())
		}
		h(&Context{c})
	}
}
