package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DarkstoneDirectory string `json:"darkstone_directory"`
}

func LoadConfig() (Config, error) {
	filePath := createConfigFilePath("user_prefs.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}

		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	return config, err
}

func SaveConfig(c Config) error {
	filePath := createConfigFilePath("user_prefs.json")

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, os.ModePerm)
}

func createConfigFilePath(fileName string) string {
	configDirectory := filepath.Join(".", "config")
	if !directoryExists(configDirectory) {
		os.MkdirAll(configDirectory, os.ModePerm)
	}

	return filepath.Join(configDirectory, fileName)
}

func directoryExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
