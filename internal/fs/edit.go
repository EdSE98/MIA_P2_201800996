package fs

import (
	"fmt"
	"os"
	"strings"

	"mia_p1_201800996/internal/disk"
)

func Edit(diskPath string, partitionStart int64, path string, contentPath string, actor Actor) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("edit requiere -path")
	}
	if strings.TrimSpace(contentPath) == "" {
		return fmt.Errorf("edit requiere -contenido")
	}

	content, err := os.ReadFile(contentPath)
	if err != nil {
		return fmt.Errorf("leer archivo de contenido: %w", err)
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
	inodeIndex, inode, err := ResolvePath(file, sb, path)
	if err != nil {
		return err
	}
	if inode.IType != '1' {
		return fmt.Errorf("la ruta no es archivo")
	}
	if !CanRead(inode, actor) || !CanWrite(inode, actor) {
		return fmt.Errorf("permiso de lectura y escritura denegado")
	}

	return WriteFileContent(file, partitionStart, &sb, inodeIndex, &inode, content)
}
