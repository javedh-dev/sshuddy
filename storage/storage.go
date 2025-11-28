package storage

import (
	"encoding/json"
	"os"

	"sshbuddy/model"
)

func GetDataPath() (string, error) {
	return "sshbuddy.json", nil
}

func LoadConfig() (*model.Config, error) {
	path, err := GetDataPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &model.Config{Hosts: []model.Host{}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config model.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(config *model.Config) error {
	path, err := GetDataPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
