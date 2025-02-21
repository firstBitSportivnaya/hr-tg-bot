package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`
	TelegramBot struct {
		Token string `yaml:"token"`
	} `yaml:"telegram_bot"`
	Database struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"dbname"`
	} `yaml:"database"`
}

func LoadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println("f.Close() failed ", err)
		}
	}(f)

	config := &Config{}
	if err := yaml.NewDecoder(f).Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
