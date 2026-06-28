package fs

import (
	"fmt"
	"os"
	"strings"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

func Rename(diskPath string, partitionStart int64, path string, newName string, actor Actor) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("rename requiere -path")
	}
	if path == "/" {
		return fmt.Errorf("no se puede renombrar la raiz")
	}
	if strings.TrimSpace(newName) == "" {
		return fmt.Errorf("el nuevo nombre no puede estar vacio")
	}
	if strings.Contains(newName, "/") {
		return fmt.Errorf("el nuevo nombre no puede contener /")
	}
	if newName == "." || newName == ".." {
		return fmt.Errorf("el nuevo nombre no puede ser . ni ..")
	}
	if len([]byte(newName)) > 12 {
		return fmt.Errorf("el nuevo nombre excede 12 bytes")
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
	parentIndex, parentInode, oldName, err := ResolveParent(file, sb, path)
	if err != nil {
		return err
	}
	targetIndex, targetInode, exists, err := FindEntryInDirectory(file, sb, parentInode, oldName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("ruta no existe: %s", path)
	}
	if !CanWrite(targetInode, actor) {
		return fmt.Errorf("permiso de escritura denegado")
	}
	if _, _, duplicate, err := FindEntryInDirectory(file, sb, parentInode, newName); err != nil {
		return err
	} else if duplicate {
		return fmt.Errorf("ya existe una entrada con nombre %q", newName)
	}

	if err := renameEntryInParent(file, sb, parentInode, oldName, targetIndex, newName); err != nil {
		return err
	}
	now := structs.NowDateBytes()
	targetInode.IMtime = now
	parentInode.IMtime = now
	if err := WriteInode(file, sb, targetIndex, targetInode); err != nil {
		return err
	}
	return WriteInode(file, sb, parentIndex, parentInode)
}

func renameEntryInParent(file *os.File, sb structs.SuperBlock, parent structs.Inode, oldName string, targetIndex int32, newName string) error {
	for i := 0; i < directBlockLimit && i < len(parent.IBlock); i++ {
		blockIndex := parent.IBlock[i]
		if blockIndex < 0 {
			continue
		}
		block, err := ReadFolderBlock(file, sb, blockIndex)
		if err != nil {
			return err
		}
		for entryIndex := range block.BContent {
			entry := &block.BContent[entryIndex]
			if entry.BInodo != targetIndex || structs.FixedBytesToString(entry.BName[:]) != oldName {
				continue
			}
			structs.SetName12(&entry.BName, newName)
			return WriteFolderBlock(file, sb, blockIndex, block)
		}
	}
	return fmt.Errorf("no se encontro la entrada a renombrar")
}
