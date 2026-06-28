package fs

import (
	"fmt"
	"os"
	"strings"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

type directorySlot struct {
	blockIndex int32
	entryIndex int
}

func Move(diskPath string, partitionStart int64, sourcePath string, destinationPath string, actor Actor) error {
	sourceParts, sourceClean, err := normalizedTransferPath(sourcePath, "move requiere -path")
	if err != nil {
		return err
	}
	if len(sourceParts) == 0 {
		return fmt.Errorf("no se puede mover la raiz")
	}
	destinationParts, _, err := normalizedTransferPath(destinationPath, "move requiere -destino")
	if err != nil {
		return err
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
	sourceParentIndex, sourceParent, name, err := ResolveParent(file, sb, sourceClean)
	if err != nil {
		return err
	}
	sourceIndex, source, exists, err := FindEntryInDirectory(file, sb, sourceParent, name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("ruta no existe: %s", sourcePath)
	}
	destinationIndex, destination, err := ResolvePath(file, sb, destinationPath)
	if err != nil {
		return err
	}
	if destination.IType != '0' {
		return fmt.Errorf("el destino no es carpeta")
	}
	if source.IType == '0' && hasPathPrefix(destinationParts, sourceParts) {
		return fmt.Errorf("no se puede mover una carpeta dentro de si misma")
	}
	if !CanWrite(source, actor) {
		return fmt.Errorf("permiso de escritura denegado")
	}
	if !CanWrite(destination, actor) {
		return fmt.Errorf("permiso de escritura denegado en destino")
	}
	if _, _, duplicate, err := FindEntryInDirectory(file, sb, destination, name); err != nil {
		return err
	} else if duplicate {
		return fmt.Errorf("ya existe una entrada con nombre %q en el destino", name)
	}

	sourceBlock, sourceEntry, err := findDirectoryEntryLocation(file, sb, sourceParent, name, sourceIndex)
	if err != nil {
		return err
	}
	destinationSlot, err := findFreeExistingDirectorySlot(file, sb, destination)
	if err != nil {
		return err
	}
	if err := writeDirectoryEntry(file, sb, destinationSlot, name, sourceIndex); err != nil {
		return err
	}
	if err := clearDirectoryEntry(file, sb, sourceBlock, sourceEntry, sourceIndex); err != nil {
		_ = clearDirectoryEntry(file, sb, destinationSlot.blockIndex, destinationSlot.entryIndex, sourceIndex)
		return err
	}
	if source.IType == '0' {
		if err := updateParentReference(file, sb, source, destinationIndex); err != nil {
			_ = writeDirectoryEntry(file, sb, directorySlot{blockIndex: sourceBlock, entryIndex: sourceEntry}, name, sourceIndex)
			_ = clearDirectoryEntry(file, sb, destinationSlot.blockIndex, destinationSlot.entryIndex, sourceIndex)
			return err
		}
	}

	now := structs.NowDateBytes()
	sourceParent.IMtime = now
	destination.IMtime = now
	source.IMtime = now
	if err := WriteInode(file, sb, sourceParentIndex, sourceParent); err != nil {
		return err
	}
	if err := WriteInode(file, sb, destinationIndex, destination); err != nil {
		return err
	}
	return WriteInode(file, sb, sourceIndex, source)
}

func normalizedTransferPath(path string, missingMessage string) ([]string, string, error) {
	if strings.TrimSpace(path) == "" {
		return nil, "", fmt.Errorf("%s", missingMessage)
	}
	parts, err := CleanAbsPath(path)
	if err != nil {
		return nil, "", err
	}
	clean := "/"
	if len(parts) > 0 {
		clean += strings.Join(parts, "/")
	}
	return parts, clean, nil
}

func hasPathPrefix(path []string, prefix []string) bool {
	if len(path) < len(prefix) {
		return false
	}
	for index := range prefix {
		if path[index] != prefix[index] {
			return false
		}
	}
	return true
}

func findFreeExistingDirectorySlot(file *os.File, sb structs.SuperBlock, directory structs.Inode) (directorySlot, error) {
	for pointerIndex := 0; pointerIndex < directBlockLimit && pointerIndex < len(directory.IBlock); pointerIndex++ {
		blockIndex := directory.IBlock[pointerIndex]
		if blockIndex < 0 {
			continue
		}
		block, err := ReadFolderBlock(file, sb, blockIndex)
		if err != nil {
			return directorySlot{}, err
		}
		for entryIndex, entry := range block.BContent {
			if entry.BInodo < 0 {
				return directorySlot{blockIndex: blockIndex, entryIndex: entryIndex}, nil
			}
		}
	}
	return directorySlot{}, fmt.Errorf("el destino no tiene una entrada libre para mover sin asignar bloques")
}

func writeDirectoryEntry(file *os.File, sb structs.SuperBlock, slot directorySlot, name string, inodeIndex int32) error {
	block, err := ReadFolderBlock(file, sb, slot.blockIndex)
	if err != nil {
		return err
	}
	if slot.entryIndex < 0 || slot.entryIndex >= len(block.BContent) || block.BContent[slot.entryIndex].BInodo >= 0 {
		return fmt.Errorf("ranura de directorio no disponible")
	}
	structs.SetName12(&block.BContent[slot.entryIndex].BName, name)
	block.BContent[slot.entryIndex].BInodo = inodeIndex
	return WriteFolderBlock(file, sb, slot.blockIndex, block)
}

func updateParentReference(file *os.File, sb structs.SuperBlock, directory structs.Inode, newParentIndex int32) error {
	for pointerIndex := 0; pointerIndex < directBlockLimit && pointerIndex < len(directory.IBlock); pointerIndex++ {
		blockIndex := directory.IBlock[pointerIndex]
		if blockIndex < 0 {
			continue
		}
		block, err := ReadFolderBlock(file, sb, blockIndex)
		if err != nil {
			return err
		}
		for entryIndex := range block.BContent {
			if structs.FixedBytesToString(block.BContent[entryIndex].BName[:]) != ".." {
				continue
			}
			block.BContent[entryIndex].BInodo = newParentIndex
			return WriteFolderBlock(file, sb, blockIndex, block)
		}
	}
	return fmt.Errorf("la carpeta no contiene referencia ..")
}
