package disk

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/structs"
)

func NormalizePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("path es obligatorio")
	}
	return filepath.Abs(path)
}

func ReadMBR(path string) (structs.MBR, error) {
	absPath, err := NormalizePath(path)
	if err != nil {
		return structs.MBR{}, err
	}
	file, err := os.OpenFile(absPath, os.O_RDONLY, 0)
	if err != nil {
		return structs.MBR{}, err
	}
	defer file.Close()

	var mbr structs.MBR
	if err := binio.ReadStructAt(file, 0, &mbr); err != nil {
		return structs.MBR{}, err
	}
	return mbr, nil
}

func WriteMBR(path string, mbr structs.MBR) error {
	absPath, err := NormalizePath(path)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(absPath, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	return binio.WriteStructAt(file, 0, mbr)
}

func OpenReadWrite(path string) (*os.File, string, error) {
	absPath, err := NormalizePath(path)
	if err != nil {
		return nil, "", err
	}
	file, err := os.OpenFile(absPath, os.O_RDWR, 0)
	if err != nil {
		return nil, "", err
	}
	return file, absPath, nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func SizeOfMBR() int64 {
	size := binary.Size(structs.MBR{})
	if size < 0 {
		return 0
	}
	return int64(size)
}

func UnitMultiplier(unit string, defaultUnit string) (int64, error) {
	normalized := strings.ToUpper(strings.TrimSpace(unit))
	if normalized == "" {
		normalized = defaultUnit
	}

	switch normalized {
	case "B":
		return 1, nil
	case "K":
		return 1024, nil
	case "M":
		return 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unidad invalida %q", unit)
	}
}

func FitToByte(fit string, defaultFit string) (byte, error) {
	normalized := strings.ToUpper(strings.TrimSpace(fit))
	if normalized == "" {
		normalized = defaultFit
	}

	switch normalized {
	case "BF", "B":
		return 'B', nil
	case "FF", "F":
		return 'F', nil
	case "WF", "W":
		return 'W', nil
	default:
		return 0, fmt.Errorf("fit invalido %q", fit)
	}
}
