package disk

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/structs"
)

type MakeDiskOptions struct {
	Size int64
	Unit string
	Fit  string
	Path string
}

func MakeDiskFromParams(params map[string]string) error {
	sizeText, ok := params["size"]
	if !ok {
		return fmt.Errorf("mkdisk requiere -size")
	}
	path, ok := params["path"]
	if !ok {
		return fmt.Errorf("mkdisk requiere -path")
	}
	size, err := strconv.ParseInt(sizeText, 10, 64)
	if err != nil {
		return fmt.Errorf("size invalido %q", sizeText)
	}

	return MakeDisk(MakeDiskOptions{
		Size: size,
		Unit: params["unit"],
		Fit:  params["fit"],
		Path: path,
	})
}

func MakeDisk(opts MakeDiskOptions) error {
	if opts.Size <= 0 {
		return fmt.Errorf("size debe ser mayor que 0")
	}

	absPath, err := NormalizePath(opts.Path)
	if err != nil {
		return err
	}
	ext := strings.ToLower(filepath.Ext(absPath))
	if ext != ".mia" && ext != ".dsk" {
		return fmt.Errorf("extension invalida %q: se acepta .mia o .dsk", ext)
	}
	if FileExists(absPath) {
		return fmt.Errorf("el disco ya existe: %s", absPath)
	}

	multiplier, err := UnitMultiplier(opts.Unit, "M")
	if err != nil {
		return err
	}
	totalSize := opts.Size * multiplier
	if totalSize <= SizeOfMBR() {
		return fmt.Errorf("el disco es demasiado pequeño para almacenar el MBR")
	}

	fit, err := FitToByte(opts.Fit, "FF")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := file.Truncate(totalSize); err != nil {
		return err
	}

	mbr := structs.NewEmptyMBR()
	mbr.MbrTamano = int32(totalSize)
	mbr.MbrFechaCreacion = structs.NowDateBytes()
	mbr.MbrDskSignature = randomSignature()
	mbr.DskFit = fit

	if err := binio.WriteStructAt(file, 0, mbr); err != nil {
		return err
	}

	size, err := binio.FileSize(file)
	if err != nil {
		return err
	}
	if size != totalSize {
		return fmt.Errorf("tamaño inesperado del disco: esperado=%d obtenido=%d", totalSize, size)
	}
	return nil
}

func randomSignature() int32 {
	var buffer [4]byte
	if _, err := rand.Read(buffer[:]); err != nil {
		return 201800996
	}
	value := int32(binary.LittleEndian.Uint32(buffer[:]))
	if value < 0 {
		value = -value
	}
	return value
}
