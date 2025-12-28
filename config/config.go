package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port    int
	Env     string
	Version string
	DBPath  string
}

func New() (*Config, error) {
	port, err := strconv.Atoi(os.Getenv("API_TODO_PORT"))
	if err != nil {
		return nil, fmt.Errorf("error parsing port: %w", err)
	}

	dbPath := os.Getenv("API_TODO_DB_PATH")
	if dbPath == "" {
		dbPath = "todo.db"
	}

	env := os.Getenv("API_TODO_ENV")
	if env == "" {
		env = "development"
	}

	version := os.Getenv("API_TODO_VERSION")

	return &Config{
		Port:    port,
		Env:     env,
		Version: version,
		DBPath:  dbPath,
	}, nil
}
