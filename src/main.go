package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/zyxar/argo/rpc"
	"golang.org/x/sync/semaphore"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const Version = "1.3.1"

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
	fmt.Println(fmt.Sprintf("\033[1;33m        \\/_/                                                          go_build.%s by 花生酱啊\033[0m", Version))
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
				if t := getToken(c); t == "" || !VerifyToken(t, config) {
					//如果token有效就不更新了，以免url改变播放器不能记住播放进度
					token, err := GenerateToken(user.Name, config)
					if err == nil {
						c.SetCookie("token", token, int(3600*config.JWT.DurationHours), "/", "", false, false)
					} else {
						log.Println("userLogin", "GenerateToken", err)
					}
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

func ass2Srt(c *gin.Context) {
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
		var userLevelBookmark = path.Join(BookmarkCacheDir, user.UserName)
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
	jsonRPC, err := rpc.New(context.Background(), config.Aria2.RPC, config.Aria2.Token, time.Second*10, &rpc.DummyNotifier{})
	if err != nil {
		log.Println("addRemoteDownloadTask", "连接Aria2失败", err)
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	defer func(jsonRPC rpc.Client) {
		_ = jsonRPC.Close()
	}(jsonRPC)
	out := c.PostForm("out")
	url := c.PostForm("url")
	seasonName, ok := getSeasonName(out)
	if !ok {
		seasonName = "Download"
	}
	dir, _ := filepath.Abs(path.Join(Root, diskManager.getMaxAvailableDisk(seasonName), seasonName))
	g, err := jsonRPC.AddURI([]string{url}, gin.H{
		"out":    out,
		"dir":    dir,
		"header": "User-Agent:" + config.Aria2.UA + "\nAccept-Encoding:identity\nConnection:Keep-Alive",
	})
	if err != nil {
		log.Println("addRemoteDownloadTask", "提交任务失败", err)
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	c.String(http.StatusOK, "%s <!-- %s -->", "<script>window.close();</script>", g)
}

func getVideoPreview(c *gin.Context) {
	_path := c.Query("path")
	previewFile := path.Join(PreviewCacheDir, path.Base(_path)+".jpg")
	if !PathExists(previewFile) {
		err := videoPreviewLock.Acquire(context.Background(), 1)
		defer videoPreviewLock.Release(1)
		if err != nil {
			log.Println("getVideoPreview", "获取锁失败", err)
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		cmd := exec.Command("ffmpeg",
			"-i", path.Join(Root, _path), "-ss",
			"00:00:05.000", "-vframes",
			"1",
			previewFile,
		)
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			log.Println("getVideoPreview", "调用ffmpeg失败", err)
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
	}
	if isAllowedPath(previewFile, PreviewCacheDir) {
		c.File(previewFile)
	} else {
		c.AbortWithStatus(http.StatusForbidden)
	}
}

func toggleBookmark(c *gin.Context) {
	_path := c.Query("path")
	user, errClaims := ParseToken(getToken(c), config)
	if errClaims != nil {
		log.Println("toggleBookmark", "从Token获取用户失败", errClaims)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	var userLevelBookmark = path.Join(BookmarkCacheDir, user.UserName)
	if !PathExists(userLevelBookmark) {
		_ = os.MkdirAll(userLevelBookmark, 0777)
	}
	var bookmarkFlagFile = path.Join(userLevelBookmark, path.Base(_path)+".b")
	state := PathExists(bookmarkFlagFile)
	if state {
		_ = os.Remove(bookmarkFlagFile)
	} else {
		fp, _ := os.Create(bookmarkFlagFile)
		defer func(fp *os.File) {
			_ = fp.Close()
		}(fp)
	}
}

// todo
func getDeviceInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"temp": 39.9,
		"fan":  true,
	})
}

var config Config
var diskManager DiskManager
var videoPreviewLock = semaphore.NewWeighted(4)

func main() {
	printLogo()
	config = *newConfig("config.json")
	diskManager = *NewDiskManager(config.MountPoints)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	authorized := router.Group("/", JWTAuth())
	//router.LoadHTMLFiles(config.WebPath+"/index.html", config.WebPath+"/login.html")
	/* router */
	router.GET("/login", sendLoginHtml)
	router.POST("/remote_download", addRemoteDownloadTask)
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
	authorized.GET("/getVideoPreview", getVideoPreview)
	authorized.GET("/toggleBookmark", toggleBookmark)
	authorized.GET("/getDeviceInfo", getDeviceInfo)
	router.POST("/uploadAss", ass2Srt)
	//router.GET("/downloadSrt", ass2Srt)
	/* router end */
	if err := router.Run(":" + strconv.Itoa(config.Port)); err != nil {
		log.Fatal("Starting NAS Failed: ", err)
	}
}
