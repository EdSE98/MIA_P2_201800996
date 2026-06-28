package fs

import (
	"fmt"
	"os"
	"strings"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

func Remove(diskPath string, partitionStart int64, path string, actor Actor) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("remove requiere -path")
	}
	parts, err := CleanAbsPath(path)
	if err != nil {
		return err
	}
	if len(parts) == 0 {
		return fmt.Errorf("no se puede eliminar la raiz")
	}

	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		return err
	}
	defer file.Close()

	sb, err := ReadSuperBlock(file, partitionStart)
	if err != nil {
		return err
	}
	parentIndex, parentInode, name, err := ResolveParent(file, sb, path)
	if err != nil {
		return err
	}
	targetIndex, targetInode, exists, err := FindEntryInDirectory(file, sb, parentInode, name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("ruta no existe: %s", path)
	}
	parentBlock, parentEntry, err := findDirectoryEntryLocation(file, sb, parentInode, name, targetIndex)
	if err != nil {
		return err
	}

	if err := removeNode(file, &sb, targetIndex, targetInode, actor, parentBlock, parentEntry); err != nil {
		return err
	}
	parentInode.IMtime = structs.NowDateBytes()
	if err := releaseEmptyDirectoryBlocks(file, &sb, &parentInode); err != nil {
		return err
	}
	return WriteInode(file, sb, parentIndex, parentInode)
}

func removeNode(file *os.File, sb *structs.SuperBlock, inodeIndex int32, inode structs.Inode, actor Actor, parentBlockIndex int32, parentEntryIndex int) error {
	if !CanWrite(inode, actor) {
		return fmt.Errorf("permiso de escritura denegado")
	}

	switch inode.IType {
	case '0':
		for pointerIndex := 0; pointerIndex < directBlockLimit && pointerIndex < len(inode.IBlock); pointerIndex++ {
			blockIndex := inode.IBlock[pointerIndex]
			if blockIndex < 0 {
				continue
			}
			block, err := ReadFolderBlock(file, *sb, blockIndex)
			if err != nil {
				return err
			}
			for entryIndex, entry := range block.BContent {
				if entry.BInodo < 0 {
					continue
				}
				entryName := structs.FixedBytesToString(entry.BName[:])
				if entryName == "." || entryName == ".." {
					continue
				}
				child, err := ReadInode(file, *sb, entry.BInodo)
				if err != nil {
					return err
				}
				if err := removeNode(file, sb, entry.BInodo, child, actor, blockIndex, entryIndex); err != nil {
					return err
				}
				inode.IMtime = structs.NowDateBytes()
				if err := WriteInode(file, *sb, inodeIndex, inode); err != nil {
					return err
				}
			}
		}
	case '1':
	default:
		return fmt.Errorf("tipo de inodo invalido")
	}

	if err := clearDirectoryEntry(file, *sb, parentBlockIndex, parentEntryIndex, inodeIndex); err != nil {
		return err
	}
	if inode.IType == '1' {
		if err := freeFileResources(file, sb, inode); err != nil {
			return err
		}
	} else {
		for pointerIndex := 0; pointerIndex < directBlockLimit && pointerIndex < len(inode.IBlock); pointerIndex++ {
			if inode.IBlock[pointerIndex] < 0 {
				continue
			}
			if err := FreeBlock(file, sb, inode.IBlock[pointerIndex]); err != nil {
				return err
			}
		}
	}
	if err := WriteInode(file, *sb, inodeIndex, structs.NewEmptyInode()); err != nil {
		return err
	}
	return FreeInode(file, sb, inodeIndex)
}

func freeFileResources(file *os.File, sb *structs.SuperBlock, inode structs.Inode) error {
	dataBlocks, err := FileDataBlockIndices(file, *sb, inode)
	if err != nil {
		return err
	}
	for _, blockIndex := range dataBlocks {
		if err := FreeBlock(file, sb, blockIndex); err != nil {
			return err
		}
	}
	if simpleIndirectIndex < len(inode.IBlock) && inode.IBlock[simpleIndirectIndex] >= 0 {
		if err := FreeBlock(file, sb, inode.IBlock[simpleIndirectIndex]); err != nil {
			return err
		}
	}
	return nil
}

func findDirectoryEntryLocation(file *os.File, sb structs.SuperBlock, parent structs.Inode, name string, inodeIndex int32) (int32, int, error) {
	for pointerIndex := 0; pointerIndex < directBlockLimit && pointerIndex < len(parent.IBlock); pointerIndex++ {
		blockIndex := parent.IBlock[pointerIndex]
		if blockIndex < 0 {
			continue
		}
		block, err := ReadFolderBlock(file, sb, blockIndex)
		if err != nil {
			return -1, -1, err
		}
		for entryIndex, entry := range block.BContent {
			if entry.BInodo == inodeIndex && structs.FixedBytesToString(entry.BName[:]) == name {
				return blockIndex, entryIndex, nil
			}
		}
	}
	return -1, -1, fmt.Errorf("no se encontro la entrada a eliminar")
}

func clearDirectoryEntry(file *os.File, sb structs.SuperBlock, blockIndex int32, entryIndex int, inodeIndex int32) error {
	block, err := ReadFolderBlock(file, sb, blockIndex)
	if err != nil {
		return err
	}
	if entryIndex < 0 || entryIndex >= len(block.BContent) || block.BContent[entryIndex].BInodo != inodeIndex {
		return fmt.Errorf("entrada de directorio inconsistente")
	}
	block.BContent[entryIndex].BName = [12]byte{}
	block.BContent[entryIndex].BInodo = -1
	return WriteFolderBlock(file, sb, blockIndex, block)
}

func releaseEmptyDirectoryBlocks(file *os.File, sb *structs.SuperBlock, inode *structs.Inode) error {
	for pointerIndex := 1; pointerIndex < directBlockLimit && pointerIndex < len(inode.IBlock); pointerIndex++ {
		blockIndex := inode.IBlock[pointerIndex]
		if blockIndex < 0 {
			continue
		}
		block, err := ReadFolderBlock(file, *sb, blockIndex)
		if err != nil {
			return err
		}
		empty := true
		for _, entry := range block.BContent {
			if entry.BInodo >= 0 {
				empty = false
				break
			}
		}
		if !empty {
			continue
		}
		if err := FreeBlock(file, sb, blockIndex); err != nil {
			return err
		}
		inode.IBlock[pointerIndex] = -1
	}
	return nil
}
