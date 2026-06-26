package services

import (
	"fmt"
	"strings"

	"mia_p1_201800996/internal/explorer"
)

func ListFS(id string, path string) (explorer.DirectoryListing, error) {
	if err := validateExplorerQuery(id, path); err != nil {
		return explorer.DirectoryListing{}, err
	}
	return explorer.List(id, path)
}

func ReadFS(id string, path string) (explorer.FileContent, error) {
	if err := validateExplorerQuery(id, path); err != nil {
		return explorer.FileContent{}, err
	}
	return explorer.Read(id, path)
}

func StatFS(id string, path string) (explorer.Metadata, error) {
	if err := validateExplorerQuery(id, path); err != nil {
		return explorer.Metadata{}, err
	}
	return explorer.Stat(id, path)
}

func validateExplorerQuery(id string, path string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("id es obligatorio")
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path es obligatorio")
	}
	return nil
}
