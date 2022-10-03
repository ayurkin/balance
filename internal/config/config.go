package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	HttpPort         string `split_words:"true"`
	PostgresUser     string `split_words:"true"`
	PostgresPassword string `split_words:"true"`
	PostgresHost     string `split_words:"true"`
	PostgresPort     string `split_words:"true"`
	PostgresDb       string `split_words:"true"`
}

func NewConfig() (*Config, error) {
	var s Config

	err := envconfig.Process("", &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}
