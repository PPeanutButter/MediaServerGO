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

type TVShow struct {
	//Adult        bool   `json:"adult"`
	//BackdropPath string `json:"backdrop_path"`
	//CreatedBy    []struct {
	//	ID          int    `json:"id"`
	//	CreditID    string `json:"credit_id"`
	//	Name        string `json:"name"`
	//	Gender      int    `json:"gender"`
	//	ProfilePath string `json:"profile_path"`
	//} `json:"created_by"`
	//EpisodeRunTime []int  `json:"episode_run_time"`
	//FirstAirDate   string `json:"first_air_date"`
	//Genres         []struct {
	//	ID   int    `json:"id"`
	//	Name string `json:"name"`
	//} `json:"genres"`
	//Homepage         string   `json:"homepage"`
	//ID               int      `json:"id"`
	//InProduction     bool     `json:"in_production"`
	//Languages        []string `json:"languages"`
	//LastAirDate      string   `json:"last_air_date"`
	//LastEpisodeToAir struct {
	//	ID             int     `json:"id"`
	//	Name           string  `json:"name"`
	//	Overview       string  `json:"overview"`
	//	VoteAverage    float64 `json:"vote_average"`
	//	VoteCount      int     `json:"vote_count"`
	//	AirDate        string  `json:"air_date"`
	//	EpisodeNumber  int     `json:"episode_number"`
	//	EpisodeType    string  `json:"episode_type"`
	//	ProductionCode string  `json:"production_code"`
	//	Runtime        int     `json:"runtime"`
	//	SeasonNumber   int     `json:"season_number"`
	//	ShowID         int     `json:"show_id"`
	//	StillPath      string  `json:"still_path"`
	//} `json:"last_episode_to_air"`
	Name  string `json:"name"`
	Title string `json:"title"`
	//NextEpisodeToAir interface{} `json:"next_episode_to_air"`
	//Networks         []struct {
	//	ID            int    `json:"id"`
	//	LogoPath      string `json:"logo_path"`
	//	Name          string `json:"name"`
	//	OriginCountry string `json:"origin_country"`
	//} `json:"networks"`
	//NumberOfEpisodes    int      `json:"number_of_episodes"`
	//NumberOfSeasons     int      `json:"number_of_seasons"`
	//OriginCountry       []string `json:"origin_country"`
	//OriginalLanguage    string   `json:"original_language"`
	OriginalName  string `json:"original_name"`
	OriginalTitle string `json:"original_title"`
	//Overview            string   `json:"overview"`
	//Popularity          float64  `json:"popularity"`
	//PosterPath          string   `json:"poster_path"`
	//ProductionCompanies []struct {
	//	ID            int    `json:"id"`
	//	LogoPath      string `json:"logo_path"`
	//	Name          string `json:"name"`
	//	OriginCountry string `json:"origin_country"`
	//} `json:"production_companies"`
	//ProductionCountries []struct {
	//	Iso31661 string `json:"iso_3166_1"`
	//	Name     string `json:"name"`
	//} `json:"production_countries"`
	//Seasons []struct {
	//	AirDate      string  `json:"air_date"`
	//	EpisodeCount int     `json:"episode_count"`
	//	ID           int     `json:"id"`
	//	Name         string  `json:"name"`
	//	Overview     string  `json:"overview"`
	//	PosterPath   string  `json:"poster_path"`
	//	SeasonNumber int     `json:"season_number"`
	//	VoteAverage  float64 `json:"vote_average"`
	//} `json:"seasons"`
	//SpokenLanguages []struct {
	//	EnglishName string `json:"english_name"`
	//	Iso6391     string `json:"iso_639_1"`
	//	Name        string `json:"name"`
	//} `json:"spoken_languages"`
	//Status      string  `json:"status"`
	//Tagline     string  `json:"tagline"`
	//Type        string  `json:"type"`
	VoteAverage float64 `json:"vote_average"`
	//VoteCount   int     `json:"vote_count"`
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
