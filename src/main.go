package main

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
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

func getAssets(c *gin.Context) {
	firstname := c.DefaultQuery("res", "index.hmtl")
	c.File(config.WebPath + "/" + firstname)
}

func getFileList(c *gin.Context) {
	path := c.Query("path")
	//todo filepath.Abs(path)，对这个结果做验证，防止目录穿越
	dirs, err := os.ReadDir(path)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}
	result := make([]gin.H, len(dirs))
	for index, dir := range dirs {
		result[index] = gin.H{
			"name": dir.Name(),
		}
	}
	c.JSON(http.StatusOK, result)
}

var config Config

func main() {
	printLogo()
	config = LoadConfig("config.json")
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.LoadHTMLFiles(config.WebPath+"/index.html", config.WebPath+"/login.html")
	/* router */
	router.GET("/", sendIndexHtml)
	router.GET("/getAssets", getAssets)
	/* authorized_router */
	authorized := router.Group("/", JWTAuth(config))
	authorized.GET("/getFileList", getFileList)
	/* router end */
	if err := router.Run(":" + strconv.Itoa(config.Port)); err != nil {
		log.Fatal("Starting NAS Failed: ", err)
	}
}
