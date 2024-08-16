package main

import (
	"github.com/gin-gonic/gin"
	"time"
)

func main() {

	r := gin.Default()
	r.Any("/ccc/stream", func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 允许所有域名访问，如果你想限制访问源，请替换 "*" 为特定域名
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		c.SSEvent("start", "start")
		for i := 0; i < 10; i++ {
			c.Writer.WriteString("data: SSE data\n\n")
			if i == 9 {
				c.SSEvent("end", "end")
			}
			c.Writer.Flush()
			time.Sleep(1 * time.Second)
		}
	})
	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}
