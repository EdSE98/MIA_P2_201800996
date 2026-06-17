package fs

import (
	"fmt"
	"os"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/structs"
)

func ReadInodeBitmap(file *os.File, sb structs.SuperBlock) ([]byte, error) {
	return ReadBitmap(file, int64(sb.SBmInodeStart), sb.SInodesCount)
}

func ReadBlockBitmap(file *os.File, sb structs.SuperBlock) ([]byte, error) {
	return ReadBitmap(file, int64(sb.SBmBlockStart), sb.SBlocksCount)
}

func AllocateInode(file *os.File, sb *structs.SuperBlock) (int32, error) {
	bitmap, err := ReadInodeBitmap(file, *sb)
	if err != nil {
		return -1, err
	}
	index, ok := FindFirstFree(bitmap)
	if !ok {
		return -1, fmt.Errorf("no hay inodos libres")
	}
	bitmap[index] = 1
	sb.SFreeInodesCount--
	sb.SFirstIno = RecalculateFirstFree(bitmap)
	if err := WriteBitmap(file, int64(sb.SBmInodeStart), bitmap); err != nil {
		return -1, err
	}
	if err := WriteSuperBlock(file, PartitionStartFromSuperBlock(*sb), *sb); err != nil {
		return -1, err
	}
	return index, nil
}

func AllocateBlock(file *os.File, sb *structs.SuperBlock) (int32, error) {
	bitmap, err := ReadBlockBitmap(file, *sb)
	if err != nil {
		return -1, err
	}
	index, ok := FindFirstFree(bitmap)
	if !ok {
		return -1, fmt.Errorf("no hay bloques libres")
	}
	bitmap[index] = 1
	sb.SFreeBlocksCount--
	sb.SFirstBlo = RecalculateFirstFree(bitmap)
	if err := WriteBitmap(file, int64(sb.SBmBlockStart), bitmap); err != nil {
		return -1, err
	}
	if err := WriteSuperBlock(file, PartitionStartFromSuperBlock(*sb), *sb); err != nil {
		return -1, err
	}
	return index, nil
}

func FreeInode(file *os.File, sb *structs.SuperBlock, inodeIndex int32) error {
	bitmap, err := ReadInodeBitmap(file, *sb)
	if err != nil {
		return err
	}
	if inodeIndex < 0 || inodeIndex >= int32(len(bitmap)) {
		return fmt.Errorf("inodo fuera de rango: %d", inodeIndex)
	}
	if bitmap[inodeIndex] == 1 {
		bitmap[inodeIndex] = 0
		sb.SFreeInodesCount++
	}
	sb.SFirstIno = RecalculateFirstFree(bitmap)
	return PersistSuperBlockAndBitmaps(file, PartitionStartFromSuperBlock(*sb), *sb, bitmap, nil)
}

func FreeBlock(file *os.File, sb *structs.SuperBlock, blockIndex int32) error {
	bitmap, err := ReadBlockBitmap(file, *sb)
	if err != nil {
		return err
	}
	if blockIndex < 0 || blockIndex >= int32(len(bitmap)) {
		return fmt.Errorf("bloque fuera de rango: %d", blockIndex)
	}
	if bitmap[blockIndex] == 1 {
		bitmap[blockIndex] = 0
		sb.SFreeBlocksCount++
	}
	sb.SFirstBlo = RecalculateFirstFree(bitmap)
	if err := WriteFileBlock(file, *sb, blockIndex, structs.NewEmptyFileBlock()); err != nil {
		return err
	}
	return PersistSuperBlockAndBitmaps(file, PartitionStartFromSuperBlock(*sb), *sb, nil, bitmap)
}

func PersistSuperBlockAndBitmaps(file *os.File, partitionStart int64, sb structs.SuperBlock, inodeBitmap []byte, blockBitmap []byte) error {
	if inodeBitmap != nil {
		if err := WriteBitmap(file, int64(sb.SBmInodeStart), inodeBitmap); err != nil {
			return err
		}
	}
	if blockBitmap != nil {
		if err := WriteBitmap(file, int64(sb.SBmBlockStart), blockBitmap); err != nil {
			return err
		}
	}
	return WriteSuperBlock(file, partitionStart, sb)
}

func PartitionStartFromSuperBlock(sb structs.SuperBlock) int64 {
	size, err := binio.BinarySize(structs.SuperBlock{})
	if err != nil {
		return 0
	}
	return int64(sb.SBmInodeStart) - size
}
