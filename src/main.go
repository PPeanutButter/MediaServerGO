package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

func printLogo() {
	file, e := os.Open("version.txt")
	if e != nil {
		panic(e.Error())
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err.Error())
		}
	}(file)
	var br = bufio.NewReader(file)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		fmt.Println(fmt.Sprintf("\033[1;33m%s\033[0m", string(line)))
	}
}

func sendIndexHtml(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func sendLoginHtml(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func getAssets(c *gin.Context) {
	firstname := c.DefaultQuery("res", "index.hmtl")
	c.File(config.WebPath + "/" + firstname)
}

func userLogin(c *gin.Context) {
	name, psw := c.Query("name"), c.Query("psw")
	for _, user := range config.Users {
		if name == user.Name {
			if MD5(psw) == user.Hash {
				// verification passed
				token, err := GenerateToken(user.Name, config)
				if err == nil {
					c.SetCookie("token", token, 604800, "/", "", false, false)
				}
				c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok"})
			} else {
				c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "密码错误"})
			}
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户不存在"})
}

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

func getFileList(c *gin.Context) {
	_path := c.Query("path")
	//todo filepath.Abs(_path)，对这个结果做验证，防止目录穿越
	dirs, err := os.ReadDir(_path)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}
	result := make([]gin.H, 0, len(dirs))
	for _, dir := range dirs {
		var fType string
		var length int64 = 0
		info, _ := dir.Info()
		if dir.IsDir() {
			fType = "Directory"
		} else if !strings.HasPrefix(dir.Name(), ".") {
			fType = "Attach"
			length = info.Size()
		} else if strings.HasPrefix(mime.TypeByExtension(path.Ext(dir.Name())), "video/") {
			fType = "File"
			length = info.Size()
		} else {
			continue
		}
		result = append(result, gin.H{
			"name":      dir.Name(),
			"mime_type": "application/octet-stream",
			"type":      fType,
			"length":    length,
			"desc":      info.ModTime(),
		})
	}
	c.JSON(http.StatusOK, result)
}

func getDeviceName(c *gin.Context) {
	hostname, err := os.Hostname()
	if err == nil {
		c.String(http.StatusOK, hostname)
	} else {
		c.String(http.StatusOK, "nil")
	}
}

var config Config

func main() {
	printLogo()
	config = LoadConfig("config.json")
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	authorized := router.Group("/", JWTAuth(config))
	//pathSafeV1 := authorized.Group("/", PathSafeV1(config))
	router.LoadHTMLFiles(config.WebPath+"/index.html", config.WebPath+"/login.html")
	/* router */
	router.GET("/login", sendLoginHtml)
	router.GET("/getAssets", getAssets)
	router.GET("/userLogin", userLogin)
	/* authorized_router */
	authorized.GET("/", sendIndexHtml)
	authorized.GET("/getDeviceName", getDeviceName)
	authorized.GET("/getFileList", getFileList)
	/* path_save_router v1 */
	//pathSafeV1
	/* router end */
	if err := router.Run(":" + strconv.Itoa(config.Port)); err != nil {
		log.Fatal("Starting NAS Failed: ", err)
	}
}
