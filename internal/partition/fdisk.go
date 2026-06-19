package partition

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

type CreateOptions struct {
	Size int64
	Unit string
	Path string
	Type string
	Fit  string
	Name string
}

func CreateFromParams(params map[string]string) error {
	sizeText, ok := params["size"]
	if !ok {
		return fmt.Errorf("fdisk requiere -size")
	}
	path, ok := params["path"]
	if !ok {
		return fmt.Errorf("fdisk requiere -path")
	}
	name, ok := params["name"]
	if !ok {
		return fmt.Errorf("fdisk requiere -name")
	}
	size, err := strconv.ParseInt(sizeText, 10, 64)
	if err != nil {
		return fmt.Errorf("size invalido %q", sizeText)
	}

	return Create(CreateOptions{
		Size: size,
		Unit: params["unit"],
		Path: path,
		Type: params["type"],
		Fit:  params["fit"],
		Name: name,
	})
}

func Create(opts CreateOptions) error {
	if opts.Size <= 0 {
		return fmt.Errorf("size debe ser mayor que 0")
	}
	if strings.TrimSpace(opts.Name) == "" {
		return fmt.Errorf("name es obligatorio")
	}
	if len(opts.Name) > 16 {
		return fmt.Errorf("name no puede exceder 16 caracteres")
	}

	partType := normalizePartitionType(opts.Type)
	if partType != 'P' && partType != 'E' && partType != 'L' {
		return fmt.Errorf("type invalido %q", opts.Type)
	}

	multiplier, err := disk.UnitMultiplier(opts.Unit, "K")
	if err != nil {
		return err
	}
	sizeBytes := opts.Size * multiplier
	fit, err := disk.FitToByte(opts.Fit, "WF")
	if err != nil {
		return err
	}

	file, absPath, err := disk.OpenReadWrite(opts.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	_ = absPath

	var mbr structs.MBR
	if err := binio.ReadStructAt(file, 0, &mbr); err != nil {
		return err
	}

	if nameExistsInMBR(mbr, opts.Name) {
		return fmt.Errorf("ya existe una particion con nombre %q", opts.Name)
	}
	if ext, _, ok := FindExtendedPartition(mbr); ok {
		exists, err := LogicalNameExists(file, ext, opts.Name)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("ya existe una particion logica con nombre %q", opts.Name)
		}
	}

	if partType == 'L' {
		extended, _, ok := FindExtendedPartition(mbr)
		if !ok {
			return fmt.Errorf("no existe particion extendida para crear logica")
		}
		return CreateLogicalPartition(file, extended, opts.Name, sizeBytes, fit)
	}

	if partType == 'E' && hasExtended(mbr) {
		return fmt.Errorf("ya existe una particion extendida")
	}

	if sizeBytes == int64(mbr.MbrTamano) && len(activePrimaryExtended(mbr)) == 0 {
		sizeBytes = int64(mbr.MbrTamano) - disk.SizeOfMBR()
	}

	slot := firstEmptySlot(mbr)
	if slot == -1 {
		return fmt.Errorf("no se pueden crear mas de 4 particiones primarias/extendidas")
	}

	start, err := allocatePrimaryOrExtended(mbr, sizeBytes, fit)
	if err != nil {
		return err
	}

	part := structs.NewEmptyPartition()
	part.PartStatus = '1'
	part.PartType = partType
	part.PartFit = fit
	part.PartStart = int32(start)
	part.PartSize = int32(sizeBytes)
	part.PartCorrelative = int32(slot + 1)
	structs.SetName16(&part.PartName, opts.Name)

	mbr.MbrPartitions[slot] = part
	if err := binio.WriteStructAt(file, 0, mbr); err != nil {
		return err
	}

	if partType == 'E' {
		ebr := structs.NewEmptyEBR()
		ebr.PartMount = '0'
		ebr.PartFit = fit
		if err := binio.WriteStructAt(file, int64(part.PartStart), ebr); err != nil {
			return err
		}
	}

	sizeAfter, err := binio.FileSize(file)
	if err != nil {
		return err
	}
	if sizeAfter != int64(mbr.MbrTamano) {
		return fmt.Errorf("el disco cambio de tamaño inesperadamente")
	}
	return nil
}

func activePrimaryExtended(mbr structs.MBR) []structs.Partition {
	var parts []structs.Partition
	for _, part := range mbr.MbrPartitions {
		if isActivePartition(part) {
			parts = append(parts, part)
		}
	}
	return parts
}

func SearchPartition(path string, name string) (structs.Partition, int, error) {
	mbr, err := disk.ReadMBR(path)
	if err != nil {
		return structs.Partition{}, -1, err
	}
	for i, part := range mbr.MbrPartitions {
		if isActivePartition(part) && structs.FixedBytesToString(part.PartName[:]) == name {
			return part, i, nil
		}
	}
	return structs.Partition{}, -1, fmt.Errorf("no existe particion %q", name)
}

func normalizePartitionType(value string) byte {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == "" {
		return 'P'
	}
	return normalized[0]
}

func firstEmptySlot(mbr structs.MBR) int {
	for i, part := range mbr.MbrPartitions {
		if !isActivePartition(part) {
			return i
		}
	}
	return -1
}

func hasExtended(mbr structs.MBR) bool {
	_, _, ok := FindExtendedPartition(mbr)
	return ok
}

func nameExistsInMBR(mbr structs.MBR, name string) bool {
	for _, part := range mbr.MbrPartitions {
		if isActivePartition(part) && structs.FixedBytesToString(part.PartName[:]) == name {
			return true
		}
	}
	return false
}

func RemoveDiskArtifacts(path string) error {
	return os.Remove(path)
}
