package common

import (
	"errors"
	"fmt"
	"log"

	// "log"
	// "os"
	// "strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var k = koanf.New(".")

type Config struct {
	NotesDir string
}

// Parse a `Config` object from the given file path
func ParseConfig(filePath string) (Config, error) {
	// if strings.HasPrefix(filePath, "~/") {
	// 	home := os.Getenv("HOME")
	// 	filePath = strings.Replace(filePath, "~", home, 1)
	// }

	if err := k.Load(file.Provider(filePath), yaml.Parser()); err != nil {
		return Config{}, fmt.Errorf("error loading config: %v", err)
	}

	return configFromKoanfGlobal()
}

func configFromKoanfGlobal() (Config, error) {
	notesDir := k.String("notes-dir")
	if notesDir == "" {
		log.Fatal("`notes-dir` must be in config file")
	}

	return Config{NotesDir: notesDir}, nil
}
