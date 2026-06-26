package services

import (
	"bytes"
	"fmt"
	"strings"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/session"
)

func MountPartition(req dto.MountRequest) (dto.MountResponse, error) {
	if strings.TrimSpace(req.Path) == "" {
		return dto.MountResponse{}, fmt.Errorf("path es obligatorio")
	}
	if strings.TrimSpace(req.Name) == "" {
		return dto.MountResponse{}, fmt.Errorf("name es obligatorio")
	}
	mounted, err := mount.Global.Mount(req.Path, req.Name)
	if err != nil {
		return dto.MountResponse{}, err
	}
	return mountToDTO(mounted), nil
}

func UnmountPartition(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("id es obligatorio")
	}
	if err := mount.Global.Unmount(id); err != nil {
		return err
	}
	session.ClearIfMountedID(id)
	return nil
}

func FormatPartition(req dto.MkfsRequest) error {
	if strings.TrimSpace(req.ID) == "" {
		return fmt.Errorf("id es obligatorio")
	}
	fsType := req.Type
	if strings.TrimSpace(fsType) == "" {
		fsType = "full"
	}
	var out bytes.Buffer
	return fs.Format(fs.FormatOptions{ID: req.ID, Type: fsType}, &out)
}
