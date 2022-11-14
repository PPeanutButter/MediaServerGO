package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := getToken(c)
		if len(auth) == 0 { //找不到任何token
			c.Redirect(http.StatusFound, "/login")
			return
		}
		// 校验token
		valid := VerifyToken(auth, config)
		if !valid {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		// token有效继续执行其他中间件
		c.Next()
	}
}

func getToken(c *gin.Context) string {
	auth := c.Request.Header.Get("Authorization")
	if auth != "" {
		return auth
	}
	auth = c.DefaultQuery("token", "")
	if auth != "" {
		return auth
	}
	auth, _ = c.Cookie("token")
	return auth
}
