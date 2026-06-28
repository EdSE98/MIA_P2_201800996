package fs

import (
	"fmt"
	"os"
	"strings"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

func Copy(diskPath string, partitionStart int64, sourcePath string, destinationPath string, actor Actor) ([]string, error) {
	sourceParts, sourceClean, err := normalizedTransferPath(sourcePath, "copy requiere -path")
	if err != nil {
		return nil, err
	}
	if len(sourceParts) == 0 {
		return nil, fmt.Errorf("no se permite copiar la raiz")
	}
	destinationParts, _, err := normalizedTransferPath(destinationPath, "copy requiere -destino")
	if err != nil {
		return nil, err
	}

	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sb, err := ReadSuperBlock(file, partitionStart)
	if err != nil {
		return nil, err
	}
	_, source, err := ResolvePath(file, sb, sourceClean)
	if err != nil {
		return nil, err
	}
	destinationIndex, destination, err := ResolvePath(file, sb, destinationPath)
	if err != nil {
		return nil, err
	}
	if destination.IType != '0' {
		return nil, fmt.Errorf("el destino no es carpeta")
	}
	if !CanRead(source, actor) {
		return nil, fmt.Errorf("permiso de lectura denegado")
	}
	if !CanWrite(destination, actor) {
		return nil, fmt.Errorf("permiso de escritura denegado en destino")
	}
	if source.IType == '0' && hasPathPrefix(destinationParts, sourceParts) {
		return nil, fmt.Errorf("no se puede copiar una carpeta dentro de si misma")
	}

	name := sourceParts[len(sourceParts)-1]
	if _, _, duplicate, err := FindEntryInDirectory(file, sb, destination, name); err != nil {
		return nil, err
	} else if duplicate {
		return nil, fmt.Errorf("ya existe una entrada con nombre %q en el destino", name)
	}

	warnings := make([]string, 0)
	createdIndex, err := copyNode(file, partitionStart, &sb, source, sourceClean, destinationIndex, &destination, name, actor, &warnings)
	if err != nil {
		if createdIndex >= 0 {
			_ = rollbackCopiedEntry(file, &sb, destinationIndex, name, createdIndex)
		}
		return warnings, err
	}
	return warnings, nil
}

func copyNode(file *os.File, partitionStart int64, sb *structs.SuperBlock, source structs.Inode, sourcePath string, destinationIndex int32, destination *structs.Inode, name string, actor Actor, warnings *[]string) (int32, error) {
	if source.IType == '1' {
		content, err := ReadFileContent(file, *sb, source)
		if err != nil {
			return -1, err
		}
		return copyFileNode(file, partitionStart, sb, source, destinationIndex, destination, name, content, actor)
	}
	if source.IType != '0' {
		return -1, fmt.Errorf("tipo de inodo invalido")
	}

	inodeIndex, err := CreateDirectory(file, partitionStart, sb, destinationIndex, destination, name, actor)
	if err != nil {
		return -1, err
	}
	copied, err := ReadInode(file, *sb, inodeIndex)
	if err != nil {
		return inodeIndex, err
	}
	copied.IPerm = source.IPerm
	if err := WriteInode(file, *sb, inodeIndex, copied); err != nil {
		return inodeIndex, err
	}

	entries, err := ListDirectoryEntries(file, *sb, source)
	if err != nil {
		return inodeIndex, err
	}
	for _, entry := range entries {
		if entry.Name == "." || entry.Name == ".." {
			continue
		}
		childPath := strings.TrimSuffix(sourcePath, "/") + "/" + entry.Name
		if !CanRead(entry.Inode, actor) {
			*warnings = append(*warnings, fmt.Sprintf("sin permiso de lectura: %s", childPath))
			continue
		}
		if _, err := copyNode(file, partitionStart, sb, entry.Inode, childPath, inodeIndex, &copied, entry.Name, actor, warnings); err != nil {
			return inodeIndex, err
		}
	}
	return inodeIndex, nil
}

func copyFileNode(file *os.File, partitionStart int64, sb *structs.SuperBlock, source structs.Inode, destinationIndex int32, destination *structs.Inode, name string, content []byte, actor Actor) (int32, error) {
	inodeIndex, err := AllocateInode(file, sb)
	if err != nil {
		return -1, err
	}
	inode := structs.NewEmptyInode()
	inode.IUid = actor.UID
	inode.IGid = actor.GID
	inode.IType = '1'
	inode.IAtime = structs.NowDateBytes()
	inode.ICtime = inode.IAtime
	inode.IMtime = inode.IAtime
	inode.IPerm = source.IPerm
	if err := WriteFileContent(file, partitionStart, sb, inodeIndex, &inode, content); err != nil {
		_ = WriteInode(file, *sb, inodeIndex, structs.NewEmptyInode())
		_ = FreeInode(file, sb, inodeIndex)
		return -1, err
	}
	if err := AddDirectoryEntry(file, sb, destinationIndex, destination, name, inodeIndex); err != nil {
		_ = freeFileResources(file, sb, inode)
		_ = WriteInode(file, *sb, inodeIndex, structs.NewEmptyInode())
		_ = FreeInode(file, sb, inodeIndex)
		return -1, err
	}
	return inodeIndex, nil
}

func rollbackCopiedEntry(file *os.File, sb *structs.SuperBlock, destinationIndex int32, name string, inodeIndex int32) error {
	destination, err := ReadInode(file, *sb, destinationIndex)
	if err != nil {
		return err
	}
	blockIndex, entryIndex, err := findDirectoryEntryLocation(file, *sb, destination, name, inodeIndex)
	if err != nil {
		return err
	}
	inode, err := ReadInode(file, *sb, inodeIndex)
	if err != nil {
		return err
	}
	return removeNode(file, sb, inodeIndex, inode, Actor{User: "root", UID: 1, GID: 1}, blockIndex, entryIndex)
}
