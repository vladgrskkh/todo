package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port int
}

func New() (*Config, error) {
	port, err := strconv.Atoi(os.Getenv("API_TODO_PORT"))
	if err != nil {
		return nil, fmt.Errorf("error parsing port: %w", err)
	}

	return &Config{
		Port: port,
	}, nil
}
