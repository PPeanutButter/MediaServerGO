package main

import (
	"io"
	"mime"
	"os"
	"path"
	"strings"
)

func isVideo(_path string) bool {
	return strings.HasPrefix(mime.TypeByExtension(path.Ext(_path)), "video/")
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
