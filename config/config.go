package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port   int
	DBPath string
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

	return &Config{
		Port:   port,
		DBPath: dbPath,
	}, nil
}
