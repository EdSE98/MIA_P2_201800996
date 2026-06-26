package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/session"
)

const defaultDisksDir = "/home/eduardo/mia/cali"

func DisksDir() string {
	if value := strings.TrimSpace(os.Getenv("MIA_DISKS_DIR")); value != "" {
		return value
	}
	return defaultDisksDir
}

func ListDisks() ([]dto.DiskResponse, error) {
	base := DisksDir()
	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return []dto.DiskResponse{}, nil
		}
		return nil, err
	}

	disks := make([]dto.DiskResponse, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".dsk" && ext != ".mia" {
			continue
		}
		path := filepath.Join(base, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		disks = append(disks, dto.DiskResponse{
			Name: entry.Name(),
			Path: absPath,
			Size: info.Size(),
		})
	}
	sort.Slice(disks, func(i, j int) bool {
		return disks[i].Path < disks[j].Path
	})
	return disks, nil
}

func CreateDisk(req dto.CreateDiskRequest) (dto.DiskResponse, error) {
	if strings.TrimSpace(req.Path) == "" {
		return dto.DiskResponse{}, fmt.Errorf("path es obligatorio")
	}
	if err := disk.MakeDisk(disk.MakeDiskOptions{
		Size: req.Size,
		Unit: req.Unit,
		Fit:  req.Fit,
		Path: req.Path,
	}); err != nil {
		return dto.DiskResponse{}, err
	}
	absPath, err := filepath.Abs(req.Path)
	if err != nil {
		return dto.DiskResponse{}, err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return dto.DiskResponse{}, err
	}
	return dto.DiskResponse{
		Name: filepath.Base(absPath),
		Path: absPath,
		Size: info.Size(),
	}, nil
}

func DeleteDisk(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path es obligatorio")
	}
	if err := disk.RemoveDisk(path); err != nil {
		return err
	}
	mount.Global.UnmountByDisk(path)
	session.ClearIfDiskPath(path)
	return nil
}
