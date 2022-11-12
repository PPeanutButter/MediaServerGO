package main

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	Port        int      `json:"port"`
	WebPath     string   `json:"webPath"`
	MountPoints []string `json:"mountPoints"`
	JWT         JWT      `json:"JWT"`
}

type JWT struct {
	Algorithm string `json:"algorithm"`
	Secret    string `json:"secret"`
}

func LoadConfig(Path string) Config {
	jsonFile, err := os.Open(Path)
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
	var config Config
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		panic(err)
	}
	return config
}
