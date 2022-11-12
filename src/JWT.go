package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func JWTAuth(cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		//todo 多种token获取方式
		auth := c.Request.Header.Get("Authorization")
		if len(auth) == 0 { //找不到任何token
			c.Redirect(http.StatusFound, "/")
			return
		}
		// 校验token
		valid := VerifyToken(auth, config)
		if !valid {
			c.Redirect(http.StatusFound, "/")
			return
		}
		// token有效继续执行其他中间件
		c.Next()
	}
}
