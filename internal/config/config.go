package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Url      string `json:"db_url"`
	Username string `json:"current_user_name"`
}

func (c *Config) SetUser(username string) error {

	c.Username = username

	writeError := write(*c)
	if writeError != nil {
		return writeError
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

func write(cfg Config) error {

	cfgData, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	writeErr := os.WriteFile(configFilePath, cfgData, 0666)
	if writeErr != nil {
		return writeErr
	}

	return nil
}
