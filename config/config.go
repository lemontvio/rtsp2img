package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	CameraSize int      `json:"cameraSize"`
	SavePath   string   `json:"savePath"`
	Cameras    []Camera `json:"cameras"`
}

func New(file string) (*Config, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(buf, &config); err != nil {
		return nil, err
	}

	if config.CameraSize == 0 {
		config.CameraSize = 3
	}

	if config.SavePath == "" {
		config.SavePath = "./camera/camera-sn.png"
	}

	return &config, nil
}

type Camera struct {
	Sn  string `json:"sn"`
	URL string `json:"url"`
}
