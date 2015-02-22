package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type TwitterConsumer struct {
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
}

type AccessToken struct {
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

type FavPConfig struct {
	Consumer    TwitterConsumer `json:"twitter_consumer"`
	AccessToken AccessToken     `json:"access_token"`
	WebHookURL  string          `json:"web_hook_url"`
	DbDsn       string          `json:"db_dsn"`
	TemplatePath string			`json:"template_path"`
}

func Parse(filename string) (FavPConfig, error) {
	var config FavPConfig
	jsonString, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("Failed to read config file:", err)
		return config, err
	}

	err = json.Unmarshal(jsonString, &config)
	if err != nil {
		log.Println("Failed to json unmarshal:", err)
		return config, nil
	}
	return config, nil
}
