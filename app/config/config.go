package config

import (
	"github.com/caarlos0/env/v7"
)

const (
	DB_TYPE_MYSQL = "mysql"
	DB_TYPE_NEO4J = "neo4j"
)

type Config struct {
	Database
}

type Database struct {
	Type     string `env:"DB_TYPE"`
	Name     string `env:"DB_NAME"`
	Username string `env:"DB_USERNAME"`
	Password string `env:"DB_PASSWORD"`
	Hostname string `env:"DB_HOSTNAME"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}

	err := env.Parse(cfg)
	return cfg, err
}
