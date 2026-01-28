package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Url      string `json:"url"`
	Username string `json:"username"`
}

func (c Config) SetUser() {

}

func write(cfg Config) error {

	userFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	userConfigData, err := os.ReadFile(userFilePath)
	if err != nil {
		return err
	}

	return nil
}

func getConfigFilePath() (string, error) {
	var configFileName = ".gatorconfig.json"

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(userHomeDir, configFileName)

	return fullPath, nil
}

func Read() (Config, error) {

	userFilePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	userConfigData, err := os.ReadFile(userFilePath)
	if err != nil {
		return Config{}, err
	}

	var userConfig Config
	if err := json.Unmarshal(userConfigData, &userConfig); err != nil {
		return Config{}, err
	}

	return userConfig, nil
}
