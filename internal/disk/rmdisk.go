package disk

import (
	"fmt"
	"os"
)

func RemoveDisk(path string) error {
	absPath, err := NormalizePath(path)
	if err != nil {
		return err
	}
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("el disco no existe: %s", absPath)
		}
		return err
	}
	return os.Remove(absPath)
}
