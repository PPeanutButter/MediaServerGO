package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
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
	fmt.Println(fmt.Sprintf("\033[1;33m        \\/_/   "+
		"                                                       go_build.%s by 花生酱啊\033[0m", Version))
}

func sendIndexHtml(c *gin.Context) {
	disableCache(c)
	c.HTML(http.StatusOK, "index.html", nil)
}

func sendLoginHtml(c *gin.Context) {
	disableCache(c)
	c.HTML(http.StatusOK, "login.html", nil)
}

func getFileCore(c *gin.Context, parent string, relative string) {
	if isAllowedPath(path.Join(parent, relative), parent) {
		c.FileAttachment(path.Join(parent, relative), path.Base(relative))
	} else {
		c.AbortWithStatus(http.StatusForbidden)
	}
}

func getFile(c *gin.Context) {
	getFileCore(c, Root, c.Query("path"))
}

func getFileV2(c *gin.Context) {
	decoded, err := base64.URLEncoding.DecodeString(c.Query("path"))
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
	}
	getFileCore(c, Root, string(decoded))
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

func getFileList(c *gin.Context) {
	_path := c.Query("path")
	user, errClaims := ParseToken(getToken(c), config)
	if errClaims != nil {
		c.AbortWithStatus(http.StatusForbidden)
	}
	if _path == "/" && len(diskManager.MountPoints) > 1 {
		//异步唤醒全部磁盘，按顺序唤醒防止启动瞬间电流过大
		go func() {
			for _, disk := range diskManager.MountPoints {
				go func(diskRoot string) {
					err := os.WriteFile(path.Join(diskRoot, ".wake"), []byte{0}, 0777)
					if err != nil {
						fmt.Printf("文件打开失败=%v\n", err)
					}
				}(disk)
				time.Sleep(time.Duration(500) * time.Millisecond)
			}
		}()
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
			if PathExists(path.Join(Root, dir, ".allow")) {
				var allowUser []string
				_ = json.Unmarshal(readBytes(path.Join(Root, dir, ".allow")), &allowUser)
				if !in(user.UserName, allowUser) {
					continue
				}
			}
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
				var info TVShow
				err = json.Unmarshal(readBytes(path.Join(Root, dir, ".info")), &info)
				score = int64(info.VoteAverage * 10)
				if info.Name != "" {
					title = info.Name
				} else {
					title = info.Title
				}
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
	out := c.PostForm("out")
	url := c.PostForm("url")
	xunlei := c.PostForm("xunlei")
	urls, isUrls := c.GetPostFormArray("urls")
	seasonName, ok := getSeasonName(out)
	if !ok {
		seasonName, ok = getMovieName(out)
		if !ok {
			seasonName = "Download"
		}
	}
	dir, _ := filepath.Abs(path.Join(Root, diskManager.getMaxAvailableDisk(seasonName), seasonName))
	var avaUrls []string
	if isUrls {
		avaUrls = urls
	} else {
		avaUrls = []string{url}
	}
	task := make(map[string]any)
	if xunlei != "" {
		task["id"] = "$XUNLEI_FID$"
	} else {
		task["id"] = "url"
	}
	task["urls"] = avaUrls
	task["name"] = out
	task["path"] = dir
	bytesData, _ := json.Marshal(task)
	g, err := http.Post("http://localhost:6900/addUrls", "application/json;charset=utf-8", bytes.NewBuffer(bytesData))
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
	// 双重检查法
	if !PathExists(previewFile) {
		err := videoPreviewLock.Acquire(context.Background(), 1)
		defer videoPreviewLock.Release(1)
		if err != nil {
			log.Println("getVideoPreview", "获取锁失败", err)
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		if !PathExists(previewFile) {
			cmd := exec.Command("ffmpeg",
				"-i", path.Join(Root, _path),
				"-ss", "00:00:05.000",
				"-vframes", "1",
				previewFile,
			)
			var out bytes.Buffer
			cmd.Stderr = &out
			err = cmd.Run()
			if err != nil {
				log.Println("getVideoPreview", "调用ffmpeg失败", cmd.Args)
				log.Println(out.String())
				c.AbortWithStatus(http.StatusServiceUnavailable)
				return
			}
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

func NotImplement(c *gin.Context) {
	c.String(http.StatusOK, "NotImplementYet, Please visit later.")
}

// todo
func getDeviceInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"temp":     -1,
		"cpu_temp": readStringFromCmdWithoutError(exec.Command("cat", "/sys/class/thermal/thermal_zone0/temp")),
		"fan":      true,
		"freq_max": readStringFromCmdWithoutError(exec.Command("cat", "/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq")),
		"freq_cur": readStringFromCmdWithoutError(exec.Command("cat", "/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_cur_freq")),
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
	router.LoadHTMLFiles(config.WebPath+"/index.html", config.WebPath+"/login.html")
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
	authorized.HEAD("/getFile/:name", getFile)
	authorized.HEAD("/getFile2/:name", getFileV2)
	authorized.GET("/getFile2/:name", getFileV2)
	authorized.GET("/getVideoPreview", getVideoPreview)
	authorized.GET("/toggleBookmark", toggleBookmark)
	authorized.GET("/getDeviceInfo", getDeviceInfo)
	/* a2s router */
	a2s := authorized.Group("/a2s")
	a2s.POST("/uploadAss", uploadAss)
	a2s.GET("/downloadSrt", downloadSrt)
	a2s.HEAD("/downloadSrt", downloadSrt)
	/* change_detection router */
	cd := authorized.Group("/change_detection", withUser("pan"))
	cd.GET("/getTasks", getTasks)
	cd.GET("/getRecords", getRecords)
	cd.GET("/getConfigurations", getConfigurations)
	cd.POST("/postTasks", NotImplement)
	cd.GET("/runOnce", runOnce)
	/* router end */
	args := os.Args
	port := ""
	if len(args) > 1 {
		port = args[1]
	} else {
		port = strconv.Itoa(config.Port)
	}
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Starting NAS Failed: ", err)
	}
}
