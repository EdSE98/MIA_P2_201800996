package fs

import (
	"fmt"
	"os"
	"strings"

	"mia_p1_201800996/internal/structs"
)

type PathEntry struct {
	Name       string
	InodeIndex int32
	Inode      structs.Inode
}

func CleanAbsPath(path string) ([]string, error) {
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("la ruta debe ser absoluta")
	}
	if path == "/" {
		return nil, nil
	}
	rawParts := strings.Split(path, "/")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		if part == "" {
			continue
		}
		if part == "." || part == ".." {
			return nil, fmt.Errorf("la ruta no puede contener . ni ..")
		}
		if len([]byte(part)) > 12 {
			return nil, fmt.Errorf("nombre excede 12 caracteres: %s", part)
		}
		parts = append(parts, part)
	}
	return parts, nil
}

func ResolvePath(file *os.File, sb structs.SuperBlock, path string) (int32, structs.Inode, error) {
	parts, err := CleanAbsPath(path)
	if err != nil {
		return -1, structs.Inode{}, err
	}
	currentIndex := int32(0)
	current, err := ReadInode(file, sb, currentIndex)
	if err != nil {
		return -1, structs.Inode{}, err
	}
	for _, part := range parts {
		nextIndex, nextInode, ok, err := FindEntryInDirectory(file, sb, current, part)
		if err != nil {
			return -1, structs.Inode{}, err
		}
		if !ok {
			return -1, structs.Inode{}, fmt.Errorf("ruta no existe: %s", path)
		}
		currentIndex = nextIndex
		current = nextInode
	}
	return currentIndex, current, nil
}

func ResolveParent(file *os.File, sb structs.SuperBlock, path string) (int32, structs.Inode, string, error) {
	parts, err := CleanAbsPath(path)
	if err != nil {
		return -1, structs.Inode{}, "", err
	}
	if len(parts) == 0 {
		return -1, structs.Inode{}, "", fmt.Errorf("la raiz no tiene padre")
	}
	parentPath := "/" + strings.Join(parts[:len(parts)-1], "/")
	if len(parts) == 1 {
		parentPath = "/"
	}
	parentIndex, parentInode, err := ResolvePath(file, sb, parentPath)
	if err != nil {
		return -1, structs.Inode{}, "", fmt.Errorf("no existe carpeta padre")
	}
	if parentInode.IType != '0' {
		return -1, structs.Inode{}, "", fmt.Errorf("el padre no es carpeta")
	}
	return parentIndex, parentInode, parts[len(parts)-1], nil
}

func FindEntryInDirectory(file *os.File, sb structs.SuperBlock, dirInode structs.Inode, name string) (int32, structs.Inode, bool, error) {
	if dirInode.IType != '0' {
		return -1, structs.Inode{}, false, fmt.Errorf("el inodo no es carpeta")
	}
	for i := 0; i < directBlockLimit && i < len(dirInode.IBlock); i++ {
		blockIndex := dirInode.IBlock[i]
		if blockIndex < 0 {
			continue
		}
		block, err := ReadFolderBlock(file, sb, blockIndex)
		if err != nil {
			return -1, structs.Inode{}, false, err
		}
		for _, content := range block.BContent {
			if content.BInodo < 0 {
				continue
			}
			if structs.FixedBytesToString(content.BName[:]) != name {
				continue
			}
			inode, err := ReadInode(file, sb, content.BInodo)
			if err != nil {
				return -1, structs.Inode{}, false, err
			}
			return content.BInodo, inode, true, nil
		}
	}
	return -1, structs.Inode{}, false, nil
}
