package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math"
	"mime"
	"os"
	"os/exec"
	"path"
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
	return false
}

func getSize(_path string) int64 {
	file, e := os.Stat(_path)
	if e == nil {
		return file.Size()
	}
	return -1
}

func readBytes(_path string) []byte {
	jsonFile, err := os.Open(_path)
	if err != nil {
		panic(err)
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			panic(err)
		}
	}(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)
	return byteValue
}

func isAllowedPath(_path string) bool {
	return strings.HasPrefix(_path, DiskManagerDir)
}

// cache: 改变时才写回
var bitrateCache = map[string]float64{}

func timeSeconds(_path string) float64 {
	var result float64 = -1
	if bitrateCache == nil && PathExists(BitRateCacheFile) {
		err := json.Unmarshal(readBytes(BitRateCacheFile), &bitrateCache)
		if err != nil {
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
			log.Println(err)
			return -1
		}
		_result, err := strconv.ParseFloat(out.String(), 32)
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
		return strconv.Itoa(int(math.Ceil(float64(8*getSize(_path)) / (seconds * 1024 * 1024))))
		//return str(math.ceil(8 * os.path.getsize(file_path) / (result * 1024 * 1024))) + "Mbps"
	} else {
		return ""
	}
}
