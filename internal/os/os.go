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

func CreatePath(path string) error {
	dir, _ := filepath.Split(path)

	if !FileExists(dir) {
		err := os.Mkdir(dir, 0700)
		if err != nil {
			return err
		}
	}

	_, err := os.Create(path)
	if err != nil {
		return err
	}

	return nil
}
