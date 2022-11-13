package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"io"
	"log"
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
	disableCache(c)
	c.HTML(http.StatusOK, "index.html", nil)
}

func sendLoginHtml(c *gin.Context) {
	disableCache(c)
	c.HTML(http.StatusOK, "login.html", nil)
}

func getFileCore(c *gin.Context, _path string) {
	log.Println(_path)
	if isAllowedPath(_path, Root) {
		c.FileAttachment(_path, path.Base(_path))
	} else {
		c.AbortWithStatus(http.StatusForbidden)
	}
}

func getFile(c *gin.Context) {
	getFileCore(c, path.Join(Root, c.Query("path")))
}

func getFileV2(c *gin.Context) {
	decoded, err := base64.URLEncoding.DecodeString(c.Query("path"))
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
	}
	getFileCore(c, path.Join(Root, string(decoded)))
}

func getAssets(c *gin.Context) {
	firstname := c.DefaultQuery("res", "index.hmtl")
	_path := path.Join(config.WebPath, firstname)
	if isAllowedPath(_path, config.WebPath) {
		c.File(_path)
	} else {
		c.AbortWithStatus(http.StatusForbidden)
	}
}

func getCover(c *gin.Context) {
	_path := path.Join(Root, c.Query("cover"), ".cover")
	if isAllowedPath(_path, Root) {
		c.File(_path)
	} else {
		c.AbortWithStatus(http.StatusForbidden)
	}
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

func getFileList(c *gin.Context) {
	_path := c.Query("path")
	user, errClaims := ParseToken(getToken(c), config)
	if errClaims != nil {
		c.AbortWithStatus(http.StatusForbidden)
	}
	dirs, err := diskManager.listDir(_path)
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
	}
	result := make([]gin.H, 0, len(dirs))
	for _, dir := range dirs {
		var fType string
		var length int64 = -1
		var score int64 = -1
		var title string
		var watchFlag = "watched"
		var bookmarkState = "bookmark_add"
		var userLevelBookmark = path.Join(BookmarkCacheDir, user.userName)
		var bookmarkFlagFile = path.Join(userLevelBookmark, path.Base(dir)+".b")
		file, e := os.Stat(path.Join(Root, dir))
		if e != nil || (file.IsDir() && !PathExists(path.Join(Root, dir, ".cover"))) {
			continue
		}
		if file.IsDir() {
			fType = "Directory"
			innerDirs, _ := os.ReadDir(path.Join(Root, dir))
			for _, innerDir := range innerDirs {
				if isVideo(innerDir.Name()) {
					if !PathExists(path.Join(userLevelBookmark, path.Base(innerDir.Name())+".b")) {
						watchFlag = ""
						break
					}
				}
			}
			if PathExists(path.Join(Root, dir, ".info")) {
				var info Info
				err = json.Unmarshal(readBytes(path.Join(Root, dir, ".info")), &info)
				score = int64(info.UserScoreChart)
				title = info.Title
			}
		} else if isVideo(file.Name()) {
			fType = "File"
			length = file.Size()
			if !PathExists(bookmarkFlagFile) {
				watchFlag = ""
			}
		} else if !strings.HasPrefix(file.Name(), ".") {
			fType = "Attach"
			length = file.Size()
		} else {
			continue
		}
		if PathExists(bookmarkFlagFile) {
			bookmarkState = "bookmark_added"
		}
		result = append(result, gin.H{
			"name":           dir,
			"mime_type":      "application/octet-stream",
			"type":           fType,
			"length":         length,
			"desc":           file.ModTime().Format("Mon Jan 2 15:04:05 2006"),
			"bookmark_state": bookmarkState,
			"watched":        watchFlag,
			"score":          score,
			"lasts":          timeSeconds(path.Join(Root, dir)),
			"bitrate":        bitrate(path.Join(Root, dir)),
			"title":          title,
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

func addRemoteDownloadTask(c *gin.Context) {
	//https://github.com/zyxar/argo
	//rpc, err := rpc2.New(context.Background(), "http://localhost:6800/jsonrpc", "0930", time.Second*10, &rpc2.DummyNotifier{})

}

var config Config
var diskManager DiskManager

func main() {
	printLogo()
	config = *newConfig("config.json")
	diskManager = *NewDiskManager(config.MountPoints)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	authorized := router.Group("/", JWTAuth(config))
	//pathSafeV1 := authorized.Group("/", PathSafeV1(config))
	router.LoadHTMLFiles(config.WebPath+"/index.html", config.WebPath+"/login.html")
	/* router */
	router.GET("/login", sendLoginHtml)
	router.GET("/getAssets", getAssets)
	router.GET("/userLogin", userLogin)
	router.GET("/remote_download", addRemoteDownloadTask)
	/* authorized_router */
	authorized.GET("/", sendIndexHtml)
	authorized.GET("/getDeviceName", getDeviceName)
	authorized.GET("/getFileList", getFileList)
	authorized.GET("/getCover", getCover)
	authorized.GET("/getFile/:name", getFile)
	authorized.GET("/getFile2/:name", getFileV2)
	/* path_save_router v1 */
	//pathSafeV1
	/* router end */
	if err := router.Run(":" + strconv.Itoa(config.Port)); err != nil {
		log.Fatal("Starting NAS Failed: ", err)
	}
}
