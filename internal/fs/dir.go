package fs

import (
	"fmt"
	"os"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

func Mkdir(diskPath string, partitionStart int64, path string, recursive bool, actor Actor) error {
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		return err
	}
	defer file.Close()

	sb, err := ReadSuperBlock(file, partitionStart)
	if err != nil {
		return err
	}
	parts, err := CleanAbsPath(path)
	if err != nil {
		return err
	}
	if len(parts) == 0 {
		return nil
	}

	currentIndex := int32(0)
	current, err := ReadInode(file, sb, currentIndex)
	if err != nil {
		return err
	}
	for i, part := range parts {
		nextIndex, nextInode, ok, err := FindEntryInDirectory(file, sb, current, part)
		if err != nil {
			return err
		}
		if ok {
			if nextInode.IType != '0' {
				return fmt.Errorf("ya existe un archivo con ese nombre: %s", path)
			}
			if i == len(parts)-1 {
				return fmt.Errorf("la carpeta ya existe: %s", path)
			}
			currentIndex = nextIndex
			current = nextInode
			continue
		}
		if !recursive && i != len(parts)-1 {
			return fmt.Errorf("no existe carpeta padre")
		}
		if !CanWrite(current, actor) {
			return fmt.Errorf("permiso de escritura denegado")
		}
		newIndex, err := CreateDirectory(file, partitionStart, &sb, currentIndex, &current, part, actor)
		if err != nil {
			return err
		}
		currentIndex = newIndex
		current, err = ReadInode(file, sb, currentIndex)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateDirectory(file *os.File, partitionStart int64, sb *structs.SuperBlock, parentIndex int32, parentInode *structs.Inode, name string, actor Actor) (int32, error) {
	inodeIndex, err := AllocateInode(file, sb)
	if err != nil {
		return -1, err
	}
	blockIndex, err := AllocateBlock(file, sb)
	if err != nil {
		_ = FreeInode(file, sb, inodeIndex)
		return -1, err
	}

	inode := structs.NewEmptyInode()
	inode.IUid = actor.UID
	inode.IGid = actor.GID
	inode.IType = '0'
	inode.IAtime = structs.NowDateBytes()
	inode.ICtime = inode.IAtime
	inode.IMtime = inode.IAtime
	inode.IBlock[0] = blockIndex
	structs.SetPerm(&inode.IPerm, "664")

	block := structs.NewEmptyFolderBlock()
	structs.SetName12(&block.BContent[0].BName, ".")
	block.BContent[0].BInodo = inodeIndex
	structs.SetName12(&block.BContent[1].BName, "..")
	block.BContent[1].BInodo = parentIndex

	if err := WriteInode(file, *sb, inodeIndex, inode); err != nil {
		return -1, err
	}
	if err := WriteFolderBlock(file, *sb, blockIndex, block); err != nil {
		return -1, err
	}
	if err := AddDirectoryEntry(file, sb, parentIndex, parentInode, name, inodeIndex); err != nil {
		return -1, err
	}
	return inodeIndex, nil
}

func AddDirectoryEntry(file *os.File, sb *structs.SuperBlock, dirIndex int32, dirInode *structs.Inode, name string, targetInode int32) error {
	if dirInode.IType != '0' {
		return fmt.Errorf("el destino no es carpeta")
	}
	for i := 0; i < directBlockLimit; i++ {
		blockIndex := dirInode.IBlock[i]
		if blockIndex < 0 {
			newBlockIndex, err := AllocateBlock(file, sb)
			if err != nil {
				return err
			}
			dirInode.IBlock[i] = newBlockIndex
			block := structs.NewEmptyFolderBlock()
			structs.SetName12(&block.BContent[0].BName, name)
			block.BContent[0].BInodo = targetInode
			if err := WriteFolderBlock(file, *sb, newBlockIndex, block); err != nil {
				return err
			}
			dirInode.IMtime = structs.NowDateBytes()
			return WriteInode(file, *sb, dirIndex, *dirInode)
		}
		block, err := ReadFolderBlock(file, *sb, blockIndex)
		if err != nil {
			return err
		}
		for j := range block.BContent {
			entryName := structs.FixedBytesToString(block.BContent[j].BName[:])
			if block.BContent[j].BInodo >= 0 && entryName == name {
				return fmt.Errorf("ya existe una entrada con ese nombre: %s", name)
			}
			if block.BContent[j].BInodo >= 0 {
				continue
			}
			structs.SetName12(&block.BContent[j].BName, name)
			block.BContent[j].BInodo = targetInode
			if err := WriteFolderBlock(file, *sb, blockIndex, block); err != nil {
				return err
			}
			dirInode.IMtime = structs.NowDateBytes()
			return WriteInode(file, *sb, dirIndex, *dirInode)
		}
	}
	return fmt.Errorf("directorio excede capacidad directa soportada")
}
