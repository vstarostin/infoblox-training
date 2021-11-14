package config

import "github.com/caarlos0/env/v6"

type Config struct {
	Port               int    `env:"PORT" envDefault:"8080"`
	GRPCPort           int    `env:"GRPC_PORT" envDefault:"9090"`
	DBConnectionString string `env:"DB_CONNECTION_STRING,notEmpty,unset,file"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
