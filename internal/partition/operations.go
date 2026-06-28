package partition

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

type ResizeOptions struct {
	Path string
	Name string
	Add  int64
	Unit string
}

type DeleteOptions struct {
	Path string
	Name string
	Mode string
}

func ResizeFromParams(params map[string]string) error {
	path, ok := params["path"]
	if !ok {
		return fmt.Errorf("fdisk -add requiere -path")
	}
	name, ok := params["name"]
	if !ok {
		return fmt.Errorf("fdisk -add requiere -name")
	}
	addText, ok := params["add"]
	if !ok {
		return fmt.Errorf("fdisk requiere -add")
	}
	add, err := strconv.ParseInt(addText, 10, 64)
	if err != nil {
		return fmt.Errorf("add invalido %q", addText)
	}
	return Resize(ResizeOptions{Path: path, Name: name, Add: add, Unit: params["unit"]})
}

func DeleteFromParams(params map[string]string) error {
	path, ok := params["path"]
	if !ok {
		return fmt.Errorf("fdisk -delete requiere -path")
	}
	name, ok := params["name"]
	if !ok {
		return fmt.Errorf("fdisk -delete requiere -name")
	}
	mode, ok := params["delete"]
	if !ok {
		return fmt.Errorf("fdisk requiere -delete")
	}
	return Delete(DeleteOptions{Path: path, Name: name, Mode: mode})
}

func Resize(opts ResizeOptions) error {
	if strings.TrimSpace(opts.Path) == "" {
		return fmt.Errorf("path es obligatorio")
	}
	if strings.TrimSpace(opts.Name) == "" {
		return fmt.Errorf("name es obligatorio")
	}
	if opts.Add == 0 {
		return fmt.Errorf("add debe ser distinto de 0")
	}

	multiplier, err := disk.UnitMultiplier(opts.Unit, "K")
	if err != nil {
		return err
	}
	delta, err := multiplySize(opts.Add, multiplier)
	if err != nil {
		return err
	}

	file, _, err := disk.OpenReadWrite(opts.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	var mbr structs.MBR
	if err := binio.ReadStructAt(file, 0, &mbr); err != nil {
		return err
	}

	for index, part := range mbr.MbrPartitions {
		if !isActivePartition(part) || structs.FixedBytesToString(part.PartName[:]) != opts.Name {
			continue
		}
		if err := resizePrimaryOrExtended(file, &mbr, index, delta); err != nil {
			return err
		}
		return binio.WriteStructAt(file, 0, mbr)
	}

	extended, _, ok := FindExtendedPartition(mbr)
	if !ok {
		return fmt.Errorf("no existe particion %q", opts.Name)
	}
	return resizeLogical(file, extended, opts.Name, delta)
}

func Delete(opts DeleteOptions) error {
	if strings.TrimSpace(opts.Path) == "" {
		return fmt.Errorf("path es obligatorio")
	}
	if strings.TrimSpace(opts.Name) == "" {
		return fmt.Errorf("name es obligatorio")
	}
	mode := strings.ToLower(strings.TrimSpace(opts.Mode))
	if mode != "fast" && mode != "full" {
		return fmt.Errorf("delete debe ser fast o full")
	}

	file, _, err := disk.OpenReadWrite(opts.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	var mbr structs.MBR
	if err := binio.ReadStructAt(file, 0, &mbr); err != nil {
		return err
	}

	for index, part := range mbr.MbrPartitions {
		if !isActivePartition(part) || structs.FixedBytesToString(part.PartName[:]) != opts.Name {
			continue
		}
		if mode == "full" {
			if err := zeroRange(file, int64(part.PartStart), int64(part.PartSize)); err != nil {
				return err
			}
		}
		mbr.MbrPartitions[index] = structs.NewEmptyPartition()
		return binio.WriteStructAt(file, 0, mbr)
	}

	extended, _, ok := FindExtendedPartition(mbr)
	if !ok {
		return fmt.Errorf("no existe particion %q", opts.Name)
	}
	return deleteLogical(file, extended, opts.Name, mode == "full")
}

func resizePrimaryOrExtended(file *os.File, mbr *structs.MBR, index int, delta int64) error {
	part := mbr.MbrPartitions[index]
	currentSize := int64(part.PartSize)
	if delta > math.MaxInt64-currentSize {
		return fmt.Errorf("el tamaño resultante excede el formato de particion")
	}
	newSize := currentSize + delta
	if newSize <= 0 {
		return fmt.Errorf("la particion no puede quedar con tamaño menor o igual a 0")
	}
	if newSize > math.MaxInt32 {
		return fmt.Errorf("el tamaño resultante excede el formato de particion")
	}

	if delta < 0 {
		formatted, err := hasEXT2SuperBlock(file, int64(part.PartStart))
		if err != nil {
			return err
		}
		if formatted {
			return fmt.Errorf("no se puede reducir una particion formateada EXT2")
		}
		if part.PartType == 'E' {
			if err := validateExtendedShrink(file, part, newSize); err != nil {
				return err
			}
		}
	} else {
		boundary := int64(mbr.MbrTamano)
		partEnd := int64(part.PartStart) + currentSize
		for otherIndex, other := range mbr.MbrPartitions {
			if otherIndex == index || !isActivePartition(other) {
				continue
			}
			if int64(other.PartStart) >= partEnd && int64(other.PartStart) < boundary {
				boundary = int64(other.PartStart)
			}
		}
		if partEnd+delta > boundary {
			return fmt.Errorf("no hay espacio contiguo suficiente para ampliar la particion")
		}
	}

	part.PartSize = int32(newSize)
	mbr.MbrPartitions[index] = part
	return nil
}

func resizeLogical(file *os.File, extended structs.Partition, name string, delta int64) error {
	chain, err := ReadEBRChain(file, extended)
	if err != nil {
		return err
	}
	for _, item := range chain {
		ebr := item.EBR
		if ebr.PartSize <= 0 || structs.FixedBytesToString(ebr.PartName[:]) != name {
			continue
		}

		currentSize := int64(ebr.PartSize)
		if delta > math.MaxInt64-currentSize {
			return fmt.Errorf("el tamaño resultante excede el formato de particion")
		}
		newSize := currentSize + delta
		if newSize <= 0 {
			return fmt.Errorf("la particion no puede quedar con tamaño menor o igual a 0")
		}
		if newSize > math.MaxInt32 {
			return fmt.Errorf("el tamaño resultante excede el formato de particion")
		}
		if delta < 0 {
			formatted, err := hasEXT2SuperBlock(file, int64(ebr.PartStart))
			if err != nil {
				return err
			}
			if formatted {
				return fmt.Errorf("no se puede reducir una particion formateada EXT2")
			}
		} else {
			boundary := int64(extended.PartStart) + int64(extended.PartSize)
			if ebr.PartNext >= 0 {
				boundary = int64(ebr.PartNext)
			}
			if int64(ebr.PartStart)+currentSize+delta > boundary {
				return fmt.Errorf("no hay espacio contiguo suficiente para ampliar la particion logica")
			}
		}

		ebr.PartSize = int32(newSize)
		return binio.WriteStructAt(file, item.Offset, ebr)
	}
	return fmt.Errorf("no existe particion %q", name)
}

func validateExtendedShrink(file *os.File, extended structs.Partition, newSize int64) error {
	ebrSize, err := binio.BinarySize(structs.EBR{})
	if err != nil {
		return err
	}
	if newSize < ebrSize {
		return fmt.Errorf("la extendida debe conservar espacio para su EBR")
	}

	chain, err := ReadEBRChain(file, extended)
	if err != nil {
		return err
	}
	requiredEnd := int64(extended.PartStart) + ebrSize
	for _, item := range chain {
		end := item.Offset + ebrSize
		if item.EBR.PartSize > 0 {
			end = int64(item.EBR.PartStart) + int64(item.EBR.PartSize)
		}
		if end > requiredEnd {
			requiredEnd = end
		}
	}
	if int64(extended.PartStart)+newSize < requiredEnd {
		return fmt.Errorf("la reduccion cortaria particiones logicas existentes")
	}
	return nil
}

func deleteLogical(file *os.File, extended structs.Partition, name string, full bool) error {
	chain, err := ReadEBRChain(file, extended)
	if err != nil {
		return err
	}
	targetIndex := -1
	for index, item := range chain {
		if item.EBR.PartSize > 0 && structs.FixedBytesToString(item.EBR.PartName[:]) == name {
			targetIndex = index
			break
		}
	}
	if targetIndex == -1 {
		return fmt.Errorf("no existe particion %q", name)
	}

	target := chain[targetIndex]
	targetEnd := int64(target.EBR.PartStart) + int64(target.EBR.PartSize)
	if targetIndex > 0 {
		previous := chain[targetIndex-1]
		previous.EBR.PartNext = target.EBR.PartNext
		if err := binio.WriteStructAt(file, previous.Offset, previous.EBR); err != nil {
			return err
		}
		if full {
			return zeroRange(file, target.Offset, targetEnd-target.Offset)
		}
		return nil
	}

	if len(chain) == 1 {
		if full {
			if err := zeroRange(file, target.Offset, targetEnd-target.Offset); err != nil {
				return err
			}
		}
		empty := structs.NewEmptyEBR()
		empty.PartMount = '0'
		empty.PartFit = extended.PartFit
		return binio.WriteStructAt(file, int64(extended.PartStart), empty)
	}

	next := chain[1]
	if full {
		if err := zeroRange(file, target.Offset, targetEnd-target.Offset); err != nil {
			return err
		}
	}
	if err := binio.WriteStructAt(file, int64(extended.PartStart), next.EBR); err != nil {
		return err
	}
	ebrSize, err := binio.BinarySize(structs.EBR{})
	if err != nil {
		return err
	}
	return zeroRange(file, next.Offset, ebrSize)
}

func hasEXT2SuperBlock(file *os.File, start int64) (bool, error) {
	var sb structs.SuperBlock
	if err := binio.ReadStructAt(file, start, &sb); err != nil {
		return false, err
	}
	return sb.SMagic == 0xEF53 && sb.SFilesystemType == 2, nil
}

func multiplySize(value int64, multiplier int64) (int64, error) {
	if value > 0 && value > math.MaxInt64/multiplier {
		return 0, fmt.Errorf("add excede el rango permitido")
	}
	if value < 0 && value < math.MinInt64/multiplier {
		return 0, fmt.Errorf("add excede el rango permitido")
	}
	return value * multiplier, nil
}

func zeroRange(file *os.File, offset int64, size int64) error {
	if err := binio.EnsureRange(file, offset, size); err != nil {
		return err
	}
	const chunkSize = 64 * 1024
	zeros := make([]byte, chunkSize)
	for written := int64(0); written < size; {
		remaining := size - written
		chunk := zeros
		if remaining < int64(len(chunk)) {
			chunk = zeros[:remaining]
		}
		if err := binio.WriteBytesAt(file, offset+written, chunk); err != nil {
			return err
		}
		written += int64(len(chunk))
	}
	return nil
}
