package main

import (
	"bytes"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"
)

func uploadAss(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "请求失败")
		return
	}
	fileName := file.Filename
	if err := c.SaveUploadedFile(file, fileName); err != nil {
		c.String(http.StatusBadRequest, "保存失败 Error:%s", err.Error())
		return
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			log.Println(err)
		}
	}(fileName)
	cmd := exec.Command("a2s", fileName)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err == nil {
		type T struct {
			Code  int      `json:"code"`
			File  string   `json:"file"`
			Files []string `json:"files"`
		}
		var result T
		err = json.Unmarshal([]byte(out.String()), &result)
		if err == nil {
			err := os.Rename(result.File, path.Join(Ass2SrtCacheDir, result.File))
			if err != nil {
				c.String(http.StatusBadRequest, "复制转化结果失败 Error:%s", err.Error())
				return
			}
			c.JSON(http.StatusOK, result)
		} else {
			c.String(http.StatusBadRequest, "读取结果失败 Error:%s", err.Error())
		}
	} else {
		c.String(http.StatusBadRequest, "调用失败 Error:%s", err.Error())
	}
}

func downloadSrt(c *gin.Context) {
	decoded, err := base64.URLEncoding.DecodeString(c.Query("path"))
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
	}
	getFileCore(c, Ass2SrtCacheDir, string(decoded))
	go func(name string) {
		time.Sleep(time.Duration(10) * time.Minute)
		err := os.Remove(name)
		if err != nil {
			log.Println(err)
		}
	}(path.Join(Ass2SrtCacheDir, string(decoded)))
}
