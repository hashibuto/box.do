package main

import (
	"auth/handlers/user"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()

	r.Use(
		gin.LoggerWithFormatter(
			func(param gin.LogFormatterParams) string {
				return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
					param.ClientIP,
					param.TimeStamp.Format(time.RFC1123Z),
					param.Method,
					param.Path,
					param.Request.Proto,
					param.StatusCode,
					param.Latency,
					param.Request.UserAgent(),
				)
			},
		),
	)

	r.Use(gin.Recovery())

	r.POST("/api/auth", user.HandlePost)

	r.Run(":3000")
}
