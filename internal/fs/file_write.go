package fs

import (
	"fmt"
	"os"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/structs"
)

const directBlockLimit = 12

func WriteRootUsersFile(file *os.File, sb structs.SuperBlock, content string) (structs.SuperBlock, error) {
	inode, err := ReadInode(file, sb, 1)
	if err != nil {
		return structs.SuperBlock{}, err
	}
	if inode.IType != '1' {
		return structs.SuperBlock{}, fmt.Errorf("/users.txt no es archivo")
	}

	blocks := SplitContentIntoBlocks(content)
	if len(blocks) == 0 {
		blocks = []structs.FileBlock{structs.NewEmptyFileBlock()}
	}
	if len(blocks) > directBlockLimit {
		return structs.SuperBlock{}, fmt.Errorf("users.txt excede capacidad directa soportada")
	}

	bitmap, err := ReadBitmap(file, int64(sb.SBmBlockStart), sb.SBlocksCount)
	if err != nil {
		return structs.SuperBlock{}, err
	}

	currentBlocks := UsedDirectBlocks(inode)
	needed := len(blocks)
	if needed > len(currentBlocks) {
		allocated, err := AllocateBlocks(bitmap, needed-len(currentBlocks))
		if err != nil {
			return structs.SuperBlock{}, err
		}
		currentBlocks = append(currentBlocks, allocated...)
		sb.SFreeBlocksCount -= int32(len(allocated))
	}
	if needed < len(currentBlocks) {
		toFree := currentBlocks[needed:]
		for _, blockIndex := range toFree {
			if blockIndex < 0 || blockIndex >= int32(len(bitmap)) {
				return structs.SuperBlock{}, fmt.Errorf("bloque usado fuera de rango: %d", blockIndex)
			}
			bitmap[blockIndex] = 0
			if err := WriteFileBlock(file, sb, blockIndex, structs.NewEmptyFileBlock()); err != nil {
				return structs.SuperBlock{}, err
			}
		}
		sb.SFreeBlocksCount += int32(len(toFree))
		currentBlocks = currentBlocks[:needed]
	}

	for i := range inode.IBlock {
		inode.IBlock[i] = -1
	}
	for i, blockIndex := range currentBlocks {
		inode.IBlock[i] = blockIndex
		if err := WriteFileBlock(file, sb, blockIndex, blocks[i]); err != nil {
			return structs.SuperBlock{}, err
		}
	}

	inode.ISize = int32(len(content))
	inode.IMtime = structs.NowDateBytes()
	sb.SFirstBlo = RecalculateFirstFree(bitmap)

	if err := WriteBitmap(file, int64(sb.SBmBlockStart), bitmap); err != nil {
		return structs.SuperBlock{}, err
	}
	if err := WriteInode(file, sb, 1, inode); err != nil {
		return structs.SuperBlock{}, err
	}
	if err := WriteSuperBlock(file, int64(sb.SBmInodeStart)-superBlockSize(), sb); err != nil {
		return structs.SuperBlock{}, err
	}
	return sb, nil
}

func SplitContentIntoBlocks(content string) []structs.FileBlock {
	data := []byte(content)
	if len(data) == 0 {
		return nil
	}
	blocks := make([]structs.FileBlock, 0, (len(data)+BlockSize-1)/BlockSize)
	for start := 0; start < len(data); start += BlockSize {
		end := start + BlockSize
		if end > len(data) {
			end = len(data)
		}
		block := structs.NewEmptyFileBlock()
		copy(block.BContent[:], data[start:end])
		blocks = append(blocks, block)
	}
	return blocks
}

func RecalculateFirstFree(bitmap []byte) int32 {
	index, ok := FindFirstFree(bitmap)
	if !ok {
		return -1
	}
	return index
}

func AllocateBlocks(bitmap []byte, count int) ([]int32, error) {
	if count < 0 {
		return nil, fmt.Errorf("cantidad de bloques invalida")
	}
	allocated := make([]int32, 0, count)
	for i := range bitmap {
		if bitmap[i] != 0 {
			continue
		}
		bitmap[i] = 1
		allocated = append(allocated, int32(i))
		if len(allocated) == count {
			return allocated, nil
		}
	}
	for _, index := range allocated {
		bitmap[int(index)] = 0
	}
	return nil, fmt.Errorf("no hay bloques libres suficientes")
}

func UsedDirectBlocks(inode structs.Inode) []int32 {
	var blocks []int32
	for i := 0; i < directBlockLimit && i < len(inode.IBlock); i++ {
		if inode.IBlock[i] >= 0 {
			blocks = append(blocks, inode.IBlock[i])
		}
	}
	return blocks
}

func superBlockSize() int64 {
	size, err := binio.BinarySize(structs.SuperBlock{})
	if err != nil {
		return 0
	}
	return size
}
