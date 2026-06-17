package fs

import (
	"fmt"
	"math"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/structs"
)

const BlockSize = 64

type Layout struct {
	PartitionStart int64
	PartitionSize  int64
	InodeCount     int32
	BlockCount     int32
	SuperStart     int64
	BmInodeStart   int64
	BmBlockStart   int64
	InodeStart     int64
	BlockStart     int64
	End            int64
}

func CalculateLayout(partitionStart int64, partitionSize int64) (Layout, error) {
	if partitionStart < 0 {
		return Layout{}, fmt.Errorf("inicio de particion invalido")
	}
	if partitionSize <= 0 {
		return Layout{}, fmt.Errorf("tamaño de particion invalido")
	}

	sbSize, err := binio.BinarySize(structs.SuperBlock{})
	if err != nil {
		return Layout{}, err
	}
	inodeSize, err := binio.BinarySize(structs.Inode{})
	if err != nil {
		return Layout{}, err
	}

	denominator := int64(1 + 3 + inodeSize + 3*BlockSize)
	n := (partitionSize - sbSize) / denominator
	if n < 2 {
		return Layout{}, fmt.Errorf("la particion es demasiado pequeña para EXT2 minimo")
	}
	if n > math.MaxInt32/3 {
		return Layout{}, fmt.Errorf("la particion excede los limites soportados")
	}

	blockCount := 3 * n
	bmInodeStart := partitionStart + sbSize
	bmBlockStart := bmInodeStart + n
	inodeStart := bmBlockStart + blockCount
	blockStart := inodeStart + n*inodeSize
	end := blockStart + blockCount*BlockSize
	partitionEnd := partitionStart + partitionSize
	if end > partitionEnd {
		return Layout{}, fmt.Errorf("layout EXT2 excede el rango de la particion")
	}
	if blockStart > math.MaxInt32 || inodeStart > math.MaxInt32 || bmBlockStart > math.MaxInt32 || bmInodeStart > math.MaxInt32 {
		return Layout{}, fmt.Errorf("offsets EXT2 exceden int32")
	}

	return Layout{
		PartitionStart: partitionStart,
		PartitionSize:  partitionSize,
		InodeCount:     int32(n),
		BlockCount:     int32(blockCount),
		SuperStart:     partitionStart,
		BmInodeStart:   bmInodeStart,
		BmBlockStart:   bmBlockStart,
		InodeStart:     inodeStart,
		BlockStart:     blockStart,
		End:            end,
	}, nil
}
