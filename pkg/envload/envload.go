// Package envload provides a small, dependency-free utility for loading
// environment variables from .env files.
//
// It supports common features such as:
//   - KEY=VALUE parsing
//   - Comments and blank lines
//   - Quoted values ("value" or 'value')
//   - Variable expansion using ${VAR}
//   - Optional overwrite behavior
//
// Syntax such as "export KEY=VALUE" is NOT supported.
package envload

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// varPattern matches ${VAR} style variable references used for
// variable expansion inside .env values.
var varPattern = regexp.MustCompile(`\$\{([A-Za-z0-9_]+)\}`)

// Load reads environment variables from the given file and sets them
// into the process environment.
//
// Override flag provides the behavior of overwriting existing environment variables
// with the ones from the .env file.
//
// Lines beginning with '#' and empty lines are ignored.
//
// Example .env file:
//
//	PORT=8080
//	HOST=localhost
//	URL=http://${HOST}:${PORT}
//
// Example usage:
//
//	err := envload.Load(".env")
//	if err != nil {
//	    slog.Error(err.Error())
//	}
func Load(filepath string, override bool) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("envload: unable to open file %s: %w", filepath, err)
	}
	defer func() {
		e := file.Close()
		if err != nil {
			err = fmt.Errorf("previous error: %w; close error: %w", err, e)
		} else {
			err = e
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			continue
		}

		value = strings.Trim(value, `"'`)

		value = expandVars(value)

		if !override {
			if _, ok := os.LookupEnv(key); ok {
				continue
			}
		}

		err := os.Setenv(key, value)
		if err != nil {
			return fmt.Errorf("envload: error setting %s: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("envload: error reading file: %w", err)
	}

	return nil
}

// expandVars replaces ${VAR} references in a value with the corresponding
// environment variable.
// If a referenced variable does not exist, it is replaced with
// an empty string.
func expandVars(value string) string {
	return varPattern.ReplaceAllStringFunc(value, func(match string) string {
		key := varPattern.FindStringSubmatch(match)[1]
		if v, ok := os.LookupEnv(key); ok {
			return v
		}
		return ""
	})
}
