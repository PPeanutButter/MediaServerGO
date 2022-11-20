package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := getToken(c)
		if len(auth) != 0 {
			valid := VerifyToken(auth, config)
			if valid {
				c.Next()
				return
			}
		}
		if c.FullPath() == "/" {
			c.Redirect(http.StatusFound, "/login")
		} else {
			c.AbortWithStatus(http.StatusForbidden)
		}
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
