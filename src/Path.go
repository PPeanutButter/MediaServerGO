package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"strings"
)

// PathSafeV1 只允许访问挂载点或者前端文件，否则认为目录穿越、目录遍历攻击，返回403
func PathSafeV1(allowed Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Query("path")
		abs, err := filepath.Abs(path)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		for _, allowedPath := range allowed.MountPoints {
			if strings.HasPrefix(abs, allowedPath) {
				c.Next()
				return
			}
		}
		if strings.HasPrefix(abs, allowed.WebPath) {
			c.Next()
			return
		}
		c.AbortWithStatus(http.StatusForbidden)
	}
}
