package fs

import (
	"fmt"
	"os"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/structs"
)

func InodeOffset(sb structs.SuperBlock, inodeIndex int32) int64 {
	return int64(sb.SInodeStart) + int64(inodeIndex)*int64(sb.SInodeSize)
}

func BlockOffset(sb structs.SuperBlock, blockIndex int32) int64 {
	return int64(sb.SBlockStart) + int64(blockIndex)*int64(sb.SBlockSize)
}

func ReadSuperBlock(file *os.File, partitionStart int64) (structs.SuperBlock, error) {
	var sb structs.SuperBlock
	if err := binio.ReadStructAt(file, partitionStart, &sb); err != nil {
		return structs.SuperBlock{}, err
	}
	return sb, nil
}

func WriteSuperBlock(file *os.File, partitionStart int64, sb structs.SuperBlock) error {
	return binio.WriteStructAt(file, partitionStart, sb)
}

func ReadInode(file *os.File, sb structs.SuperBlock, inodeIndex int32) (structs.Inode, error) {
	if err := validateIndex("inodo", inodeIndex, sb.SInodesCount); err != nil {
		return structs.Inode{}, err
	}
	var inode structs.Inode
	if err := binio.ReadStructAt(file, InodeOffset(sb, inodeIndex), &inode); err != nil {
		return structs.Inode{}, err
	}
	return inode, nil
}

func WriteInode(file *os.File, sb structs.SuperBlock, inodeIndex int32, inode structs.Inode) error {
	if err := validateIndex("inodo", inodeIndex, sb.SInodesCount); err != nil {
		return err
	}
	return binio.WriteStructAt(file, InodeOffset(sb, inodeIndex), inode)
}

func ReadFolderBlock(file *os.File, sb structs.SuperBlock, blockIndex int32) (structs.FolderBlock, error) {
	if err := validateIndex("bloque", blockIndex, sb.SBlocksCount); err != nil {
		return structs.FolderBlock{}, err
	}
	var block structs.FolderBlock
	if err := binio.ReadStructAt(file, BlockOffset(sb, blockIndex), &block); err != nil {
		return structs.FolderBlock{}, err
	}
	return block, nil
}

func WriteFolderBlock(file *os.File, sb structs.SuperBlock, blockIndex int32, block structs.FolderBlock) error {
	if err := validateIndex("bloque", blockIndex, sb.SBlocksCount); err != nil {
		return err
	}
	return binio.WriteStructAt(file, BlockOffset(sb, blockIndex), block)
}

func ReadFileBlock(file *os.File, sb structs.SuperBlock, blockIndex int32) (structs.FileBlock, error) {
	if err := validateIndex("bloque", blockIndex, sb.SBlocksCount); err != nil {
		return structs.FileBlock{}, err
	}
	var block structs.FileBlock
	if err := binio.ReadStructAt(file, BlockOffset(sb, blockIndex), &block); err != nil {
		return structs.FileBlock{}, err
	}
	return block, nil
}

func WriteFileBlock(file *os.File, sb structs.SuperBlock, blockIndex int32, block structs.FileBlock) error {
	if err := validateIndex("bloque", blockIndex, sb.SBlocksCount); err != nil {
		return err
	}
	return binio.WriteStructAt(file, BlockOffset(sb, blockIndex), block)
}

func ReadRootUsersFile(file *os.File, sb structs.SuperBlock) (string, error) {
	inode, err := ReadInode(file, sb, 1)
	if err != nil {
		return "", err
	}
	if inode.IType != '1' {
		return "", fmt.Errorf("/users.txt no es archivo")
	}

	remaining := int(inode.ISize)
	content := make([]byte, 0, remaining)
	for _, blockIndex := range inode.IBlock {
		if remaining <= 0 {
			break
		}
		if blockIndex < 0 {
			continue
		}
		block, err := ReadFileBlock(file, sb, blockIndex)
		if err != nil {
			return "", err
		}
		take := remaining
		if take > len(block.BContent) {
			take = len(block.BContent)
		}
		content = append(content, block.BContent[:take]...)
		remaining -= take
	}
	if remaining > 0 {
		return "", fmt.Errorf("contenido de /users.txt incompleto")
	}
	return string(content), nil
}

func validateIndex(label string, index int32, count int32) error {
	if index < 0 {
		return fmt.Errorf("indice de %s negativo: %d", label, index)
	}
	if index >= count {
		return fmt.Errorf("indice de %s fuera de rango: %d", label, index)
	}
	return nil
}
