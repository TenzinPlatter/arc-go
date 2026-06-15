package internal

import (
	"errors"
	"fmt"
	"log/slog"

	"os"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var k = koanf.New(".")

type Config struct {
	NotesDir string
	ApiToken string
}

func replaceTilde(s string) (string, error) {
	if !strings.HasPrefix(s, "~/") {
		return s, nil
	}

	home, wasSet := os.LookupEnv("HOME")
	if !wasSet {
		return "", errors.New("$HOME env var was not set, could not replace tilde")
	}

	return strings.Replace(s, "~", home, 1), nil
}

// Parse a `Config` object from the given file path
func ParseConfig(filePath string) (Config, error) {
	slog.Debug("Parsing config from " + filePath)
	filePath, err := replaceTilde(filePath)
	if err != nil {
		return Config{}, nil
	}

	if err := k.Load(file.Provider(filePath), toml.Parser()); err != nil {
		return Config{}, fmt.Errorf("error loading config: %v", err)
	}

	return configFromKoanfGlobal()
}

func configFromKoanfGlobal() (Config, error) {
	notesDir, err := replaceTilde(k.String("notes-dir"))
	if notesDir == "" {
		return Config{}, errors.New("`notes-dir` must be non empty")
	}
	if err != nil {
		return Config{}, fmt.Errorf("Failed to parse config: %s\n", err.Error())
	}

	notesDir = strings.TrimSuffix(notesDir, "/")
	
	apiToken := k.String("api-token")
	if apiToken == "" {
		return Config{}, errors.New("`api-token` config field must be set and non-empty")
	}

	return Config{NotesDir: notesDir, ApiToken: apiToken}, nil
}
