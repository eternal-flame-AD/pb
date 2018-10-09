package main

import (
	"encoding/json"
	"errors"

	"github.com/shibukawa/configdir"
)

var config Config

type Config struct {
	Key string `json:"key" key:"key" help:"Pushbullet API key"`
}

func getConfDir() (*configdir.Config, error) {
	configDirs := configdir.New(AppVendor, AppName)

	folders := configDirs.QueryFolders(configdir.Global)
	if len(folders) == 0 {
		return nil, errors.New("Failed to find a config folder")
	}

	return folders[0], nil
}

func ExistConfig() bool {
	dir, err := getConfDir()
	if err != nil {
		panic(err)
	}
	return dir.Exists("config.json")
}

func ReadConfig() error {
	dir, err := getConfDir()
	if err != nil {
		return err
	}
	data, err := dir.ReadFile("config.json")
	if err != nil {
		return err
	}
	config = Config{}
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}
	return nil
}

func WriteConfig() error {
	dir, err := getConfDir()
	if err != nil {
		return err
	}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	if err := dir.WriteFile("config.json", data); err != nil {
		return err
	}
	return nil
}
