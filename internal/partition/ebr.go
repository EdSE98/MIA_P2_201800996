package partition

import (
	"fmt"
	"os"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/structs"
)

type EBRInfo struct {
	Offset int64
	EBR    structs.EBR
}

func FindExtendedPartition(mbr structs.MBR) (structs.Partition, int, bool) {
	for i, part := range mbr.MbrPartitions {
		if isActivePartition(part) && part.PartType == 'E' {
			return part, i, true
		}
	}
	return structs.Partition{}, -1, false
}

func ReadEBRChain(file *os.File, extended structs.Partition) ([]EBRInfo, error) {
	if !isActivePartition(extended) || extended.PartType != 'E' {
		return nil, fmt.Errorf("particion extendida invalida")
	}

	var chain []EBRInfo
	offset := int64(extended.PartStart)
	limit := int64(extended.PartStart + extended.PartSize)

	for {
		if offset < int64(extended.PartStart) || offset >= limit {
			return nil, fmt.Errorf("cadena EBR fuera de rango")
		}

		var ebr structs.EBR
		if err := binio.ReadStructAt(file, offset, &ebr); err != nil {
			return nil, err
		}
		chain = append(chain, EBRInfo{Offset: offset, EBR: ebr})

		if ebr.PartNext == -1 {
			break
		}
		offset = int64(ebr.PartNext)
	}

	return chain, nil
}

func LogicalNameExists(file *os.File, extended structs.Partition, name string) (bool, error) {
	chain, err := ReadEBRChain(file, extended)
	if err != nil {
		return false, err
	}

	for _, item := range chain {
		if item.EBR.PartSize > 0 && structs.FixedBytesToString(item.EBR.PartName[:]) == name {
			return true, nil
		}
	}
	return false, nil
}

func CreateLogicalPartition(file *os.File, extended structs.Partition, name string, size int64, fit byte) error {
	ebrSize, err := binio.BinarySize(structs.EBR{})
	if err != nil {
		return err
	}
	chain, err := ReadEBRChain(file, extended)
	if err != nil {
		return err
	}

	extendedEnd := int64(extended.PartStart + extended.PartSize)
	first := chain[0]
	if first.EBR.PartSize <= 0 {
		if int64(extended.PartStart)+ebrSize+size > extendedEnd {
			return fmt.Errorf("no hay espacio suficiente dentro de la extendida")
		}
		ebr := structs.NewEmptyEBR()
		ebr.PartMount = '0'
		ebr.PartFit = fit
		ebr.PartStart = int32(int64(extended.PartStart) + ebrSize)
		ebr.PartSize = int32(size)
		ebr.PartNext = -1
		structs.SetName16(&ebr.PartName, name)
		return binio.WriteStructAt(file, int64(extended.PartStart), ebr)
	}

	last := chain[len(chain)-1]
	nextEBROffset := int64(last.EBR.PartStart) + int64(last.EBR.PartSize)
	if nextEBROffset+ebrSize+size > extendedEnd {
		return fmt.Errorf("no hay espacio suficiente dentro de la extendida")
	}

	last.EBR.PartNext = int32(nextEBROffset)
	if err := binio.WriteStructAt(file, last.Offset, last.EBR); err != nil {
		return err
	}

	ebr := structs.NewEmptyEBR()
	ebr.PartMount = '0'
	ebr.PartFit = fit
	ebr.PartStart = int32(nextEBROffset + ebrSize)
	ebr.PartSize = int32(size)
	ebr.PartNext = -1
	structs.SetName16(&ebr.PartName, name)
	return binio.WriteStructAt(file, nextEBROffset, ebr)
}
