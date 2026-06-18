package reports

import (
	"fmt"
	"os"

	"mia_p1_201800996/internal/fs"
)

func BuildFileReport(ctx Ext2Context, file *os.File, path string) (string, error) {
	_, inode, err := fs.ResolvePath(file, ctx.SuperBlock, path)
	if err != nil {
		return "", err
	}
	if inode.IType != '1' {
		return "", fmt.Errorf("la ruta no es archivo")
	}
	content, err := fs.ReadFileContent(file, ctx.SuperBlock, inode)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Archivo: %s\n====================\n%s", path, string(content)), nil
}
