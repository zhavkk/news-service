package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env   string      `yaml:"env" env:"ENV" env-default:"local"`
	HTTP  HTTPConfig  `yaml:"http"`
	DBURL string      `yaml:"db_url"`
	Redis RedisConfig `yaml:"redis"`
}

type HTTPConfig struct {
	Port         int           `yaml:"port" env:"PORT" env-default:"8080"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"5s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env-default:"5s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env-default:"120s"`
}

type RedisConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	DialTimeout  time.Duration `yaml:"dial_timeout" env-default:"5s"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"3s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env-default:"3s"`
}

func (c RedisConfig) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
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
