package os

import (
	"os"
	"path/filepath"
)

func ConfigPath() string {
	config, _ := os.UserConfigDir()
	path := filepath.Join(config, "Labee")

	return path
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func CreateFile(path string) error {
	dir, _ := filepath.Split(path)

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	return file.Close()
}
