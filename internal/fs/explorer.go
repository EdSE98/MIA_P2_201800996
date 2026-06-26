package fs

import (
	"os"

	"mia_p1_201800996/internal/structs"
)

func ListDirectoryEntries(file *os.File, sb structs.SuperBlock, dirInode structs.Inode) ([]PathEntry, error) {
	if dirInode.IType != '0' {
		return nil, ErrNotDirectory()
	}
	entries := make([]PathEntry, 0)
	for i := 0; i < directBlockLimit && i < len(dirInode.IBlock); i++ {
		blockIndex := dirInode.IBlock[i]
		if blockIndex < 0 {
			continue
		}
		block, err := ReadFolderBlock(file, sb, blockIndex)
		if err != nil {
			return nil, err
		}
		for _, content := range block.BContent {
			if content.BInodo < 0 {
				continue
			}
			name := structs.FixedBytesToString(content.BName[:])
			inode, err := ReadInode(file, sb, content.BInodo)
			if err != nil {
				return nil, err
			}
			entries = append(entries, PathEntry{
				Name:       name,
				InodeIndex: content.BInodo,
				Inode:      inode,
			})
		}
	}
	return entries, nil
}

func ErrNotDirectory() error {
	return errNotDirectory{}
}

type errNotDirectory struct{}

func (errNotDirectory) Error() string {
	return "la ruta no es carpeta"
}
