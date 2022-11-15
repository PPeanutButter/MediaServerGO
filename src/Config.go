package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type Config struct {
	Port        int      `json:"port"`
	WebPath     string   `json:"webPath"`
	MountPoints []string `json:"mountPoints"`
	JWT         JWT      `json:"JWT"`
	Users       []User   `json:"users"`
	Aria2       Aria2    `json:"Aria2"`
}

type JWT struct {
	Algorithm     string `json:"algorithm"`
	Secret        string `json:"secret"`
	DurationHours int64  `json:"durationHours"`
}

type User struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

type Info struct {
	Title          string `json:"title"`
	Certification  string `json:"certification"`
	Genres         string `json:"genres"`
	Runtime        string `json:"runtime"`
	Tagline        string `json:"tagline"`
	Overview       string `json:"overview"`
	UserScoreChart int    `json:"user_score_chart"`
	Url            string `json:"url"`
}

type Aria2 struct {
	RPC   string `json:"RPC"`
	UA    string `json:"UA"`
	Token string `json:"Token"`
}

func newConfig(Path string) *Config {
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
		log.Println("newConfig", "读取Config失败、请检查字段", err)
		panic(err)
	}
	return &config
}
