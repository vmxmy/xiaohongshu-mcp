package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	URLs struct {
		Creator struct {
			PublishImage string `yaml:"publish_image"`
			PublishVideo string `yaml:"publish_video"`
		} `yaml:"creator"`
	} `yaml:"urls"`
	Limits struct {
		MaxTags   int `yaml:"max_tags"`
		MinImages int `yaml:"min_images"`
		MaxImages int `yaml:"max_images"`
	} `yaml:"limits"`
	Timeouts struct {
		Navigate    int `yaml:"navigate"`
		ElementWait int `yaml:"element_wait"`
		ImageUpload int `yaml:"image_upload"`
	} `yaml:"timeouts"`
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
