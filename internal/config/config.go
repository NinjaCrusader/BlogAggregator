package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Url      string `json:"url"`
	Username string `json:"username"`
}

func (c Config) SetUser() {

}

func Read() Config {

	userHomeDirectory := os.UserHomeDir()

	userConfigData, err := os.ReadFile("~/.gatorconfig.json")
	if err != nil {
		log.Fatal(err)
	}

	var userConfig Config
	if err := json.Unmarshal(userConfigData, &userConfig); err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	return userConfig
}
