package envload

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempEnv(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatal("failed to write temp .env file: %w", err)
	}

	return path
}

func clearEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		t.Setenv(k, "")
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expect   map[string]string
		preload  map[string]string
		override bool
	}{
		{
			name: "basic KEY=VALUE",
			env: `
				PORT=8080
				ENV=development
				`,
			expect: map[string]string{
				"PORT": "8080",
				"ENV":  "development",
			},
			override: true,
		},
		{
			name: "comments and whitespaces",
			env: `
				# This is a comment
				# This is another comment
				PORT = 8080
				`,
			expect: map[string]string{
				"PORT": "8080",
			},
			override: true,
		},
		{
			name: "quoted values",
			env: `
				PORT="8080"
				ENV='development'
				`,
			expect: map[string]string{
				"PORT": "8080",
				"ENV":  "development",
			},
			override: true,
		},
		{
			name: "variable expansion",
			env: `
				HOST=localhost
				PORT=8080
				URL=http://${HOST}:${PORT}
				`,
			expect: map[string]string{
				"HOST": "localhost",
				"PORT": "8080",
				"URL":  "http://localhost:8080",
			},
			override: true,
		},
		{
			name: "missing variable expansion",
			env: `
				PORT=${MISSING}
				`,
			expect: map[string]string{
				"PORT": "",
			},
			override: true,
		},
		{
			name: "does not override existing variables",
			env: `
				PORT=8080
				`,
			preload: map[string]string{
				"PORT": "1234",
			},
			expect: map[string]string{
				"PORT": "1234",
			},
		},
		{
			name: "override existing variables",
			env: `
				PORT=8080
				`,
			preload: map[string]string{
				"PORT": "1234",
			},
			expect: map[string]string{
				"PORT": "8080",
			},
			override: true,
		},
		{
			name: "mailformed line skipped",
			env: `
				PORT
				ENV=development
				=smth
				`,
			expect: map[string]string{
				"ENV": "development",
			},
			override: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t, "PORT", "ENV", "HOST", "URL")
			for k, v := range tt.preload {
				t.Setenv(k, v)
			}

			path := writeTempEnv(t, tt.env)

			if err := Load(path, tt.override); err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			for k, expected := range tt.expect {
				if got := os.Getenv(k); got != expected {
					t.Errorf("%s: expected '%s', got `%s`", k, expected, got)
				}
			}
		})
	}
}
