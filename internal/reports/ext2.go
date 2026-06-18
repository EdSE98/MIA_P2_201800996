package reports

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/structs"
)

type Ext2Context struct {
	ID             string
	DiskPath       string
	PartitionStart int64
	PartitionSize  int64
	SuperBlock     structs.SuperBlock
}

func LoadExt2Context(id string) (Ext2Context, *os.File, error) {
	mounted, ok := mount.Global.GetMounted(id)
	if !ok {
		return Ext2Context{}, nil, fmt.Errorf("no existe montaje con id %q", id)
	}
	file, _, err := disk.OpenReadWrite(mounted.DiskPath)
	if err != nil {
		return Ext2Context{}, nil, err
	}
	sb, err := fs.ReadSuperBlock(file, int64(mounted.Start))
	if err != nil {
		file.Close()
		return Ext2Context{}, nil, err
	}
	if sb.SMagic != 0xEF53 || sb.SFilesystemType != 2 {
		file.Close()
		return Ext2Context{}, nil, fmt.Errorf("la particion no tiene formato EXT2")
	}
	return Ext2Context{
		ID:             mounted.ID,
		DiskPath:       mounted.DiskPath,
		PartitionStart: int64(mounted.Start),
		PartitionSize:  int64(mounted.Size),
		SuperBlock:     sb,
	}, file, nil
}

func UsedInodeIndices(file *os.File, sb structs.SuperBlock) ([]int32, error) {
	bitmap, err := ReadInodeBitmapForReport(file, sb)
	if err != nil {
		return nil, err
	}
	return usedIndices(bitmap), nil
}

func UsedBlockIndices(file *os.File, sb structs.SuperBlock) ([]int32, error) {
	bitmap, err := ReadBlockBitmapForReport(file, sb)
	if err != nil {
		return nil, err
	}
	return usedIndices(bitmap), nil
}

func ReadInodeBitmapForReport(file *os.File, sb structs.SuperBlock) ([]byte, error) {
	return fs.ReadInodeBitmap(file, sb)
}

func ReadBlockBitmapForReport(file *os.File, sb structs.SuperBlock) ([]byte, error) {
	return fs.ReadBlockBitmap(file, sb)
}

func DotOrRender(dot string, outputPath string) error {
	_, err := writeDotAndRender(outputPath, dot)
	return err
}

func writePlainText(outputPath string, content string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outputPath, []byte(content), 0o644)
}

func usedIndices(bitmap []byte) []int32 {
	var result []int32
	for i, value := range bitmap {
		if value == 1 {
			result = append(result, int32(i))
		}
	}
	return result
}

func cleanBlockText(data []byte) string {
	return strings.TrimRight(string(data), "\x00")
}
