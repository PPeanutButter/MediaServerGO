package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io"
	"log"
	"math"
	"mime"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func isVideo(_path string) bool {
	return strings.HasPrefix(mime.TypeByExtension(path.Ext(_path)), "video/")
}

func isFile(_path string) bool {
	file, e := os.Stat(_path)
	if e == nil {
		return !file.IsDir()
	}
	log.Println("isFile", "os.Stat", e)
	return false
}

func getSize(_path string) int64 {
	file, e := os.Stat(_path)
	if e == nil {
		return file.Size()
	}
	log.Println("getSize", "os.Stat", e)
	return -1
}

func readBytes(_path string) []byte {
	jsonFile, err := os.Open(_path)
	if err != nil {
		log.Println("readBytes", "os.Open", err)
		panic(err)
	}
	defer func(jsonFile *os.File) {
		_ = jsonFile.Close()
	}(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)
	return byteValue
}

func isAllowedPath(_path string, allowed string) bool {
	a, _ := filepath.Abs(_path)
	b, _ := filepath.Abs(allowed)
	return strings.HasPrefix(a, b)
}

// cache: 改变时才写回
var bitrateCache map[string]float64 = nil

func timeSeconds(_path string) float64 {
	var result float64 = -1
	if bitrateCache == nil && PathExists(BitRateCacheFile) {
		err := json.Unmarshal(readBytes(BitRateCacheFile), &bitrateCache)
		if err != nil {
			log.Println("timeSeconds", "Unmarshal", err)
			return -1
		}
	}
	if isFile(_path) && isVideo(_path) {
		rate, found := bitrateCache[_path]
		if found {
			//hit cache
			return rate
		}
		//call ffmpeg
		cmd := exec.Command("ffprobe",
			"-v", "error", "-show_entries",
			"format=duration", "-of",
			"default=noprint_wrappers=1:nokey=1",
			_path,
		)
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			log.Println("timeSeconds", "调用ffprobe", err)
			return -1
		}
		_result, err := strconv.ParseFloat(strings.Trim(strings.Trim(out.String(), "\n"), "\r\n"), 64)
		result = _result
		//更新内存缓存
		bitrateCache[_path] = _result
		//异步刷回磁盘
		go func() {
			filePtr, _ := os.Create(BitRateCacheFile)
			defer func(filePtr *os.File) {
				_ = filePtr.Close()
			}(filePtr)
			// 创建Json编码器
			encoder := json.NewEncoder(filePtr)
			_ = encoder.Encode(bitrateCache)
		}()
	}
	return result
}

func bitrate(_path string) string {
	seconds := timeSeconds(_path)
	if seconds > 0 {
		return strconv.Itoa(int(math.Ceil(float64(8*getSize(_path))/(seconds*1024*1024)))) + "Mbps"
	} else {
		return ""
	}
}

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

func disableCache(c *gin.Context) {
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "public, max-age=0")
}

func getSeasonName(_path string) (string, bool) {
	var re = regexp.MustCompile(`(?mi)(.*)\.S\d{2}E\d{2}\.`)
	match := re.FindStringSubmatch(_path)
	if match != nil {
		return ToCamelCase(match[1]), true
	}
	return "", false
}

func getMovieName(_path string) (string, bool) {
	var re = regexp.MustCompile(`(?mi)(.*)\.(18|19|20)\d{2}\.`)
	match := re.FindStringSubmatch(_path)
	if match != nil {
		return ToCamelCase(match[1]), true
	}
	return "", false
}

func ToCamelCase(s string) string {
	s = strings.ReplaceAll(s, ".", " ")
	return cases.Title(language.Und).String(s)
}

func readStringFromCmd(cmd *exec.Cmd) (string, error) {
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err == nil {
		return out.String(), nil
	}
	return out.String(), err
}

func readStringFromCmdWithoutError(cmd *exec.Cmd) string {
	result, err := readStringFromCmd(cmd)
	if err == nil {
		return result
	}
	return ""
}
