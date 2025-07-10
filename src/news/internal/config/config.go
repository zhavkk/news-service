package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env  string `yaml:"env" env:"ENV" env-default:"local"`
	Port int    `yaml:"port" env:"PORT" env-default:"8080"`
}

func MustLoad(configPath string) *Config {
	if configPath == "" {
		panic("Config path is not set")
	}
	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("config file does not exist: %s", configPath)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("can't read config %s", err.Error())
	}
	return &cfg
}
