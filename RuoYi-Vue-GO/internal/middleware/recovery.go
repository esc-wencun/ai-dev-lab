package middleware

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"ruoyi-vue-go/pkg/response"
)

// Recovery panic 兜底：转 code=500 信封（HTTP 恒 200），替代 gin.Recovery 的纯 HTTP 500。
// 消息对位 Java/Python 泛化异常处理：输出异常文本（panic value 为 error 时取 Error()）。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				msg := fmt.Sprint(rec)
				if err, ok := rec.(error); ok {
					msg = err.Error()
				}
				log.Printf("请求地址'%s',发生系统异常: %s", c.Request.URL.Path, msg)
				response.Error(c, msg).JSON()
			}
		}()
		c.Next()
	}
}
