package fs

import (
	"fmt"
	"io"
	"os"
	"strings"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/structs"
)

const initialUsersContent = "1,G,root\n1,U,root,root,123\n"

type FormatOptions struct {
	ID   string
	Type string
}

func FormatFromParams(params map[string]string, out io.Writer) error {
	id, ok := params["id"]
	if !ok || strings.TrimSpace(id) == "" {
		return fmt.Errorf("mkfs requiere -id")
	}
	fsType := params["type"]
	if strings.TrimSpace(fsType) == "" {
		fsType = "full"
	}
	return Format(FormatOptions{ID: id, Type: fsType}, out)
}

func Format(opts FormatOptions, out io.Writer) error {
	if strings.ToLower(strings.TrimSpace(opts.Type)) != "full" {
		return fmt.Errorf("tipo de mkfs invalido %q: solo se permite full", opts.Type)
	}

	mounted, ok := mount.Global.GetMounted(opts.ID)
	if !ok {
		return fmt.Errorf("no existe montaje con id %q", opts.ID)
	}
	if mounted.Start < 0 || mounted.Size <= 0 {
		return fmt.Errorf("rango de particion invalido para %s", opts.ID)
	}
	if mounted.PartitionType == 'E' {
		fmt.Fprintln(out, "Advertencia: formateando particion extendida por compatibilidad con script")
	}

	file, _, err := disk.OpenReadWrite(mounted.DiskPath)
	if err != nil {
		return err
	}
	defer file.Close()

	partStart := int64(mounted.Start)
	partSize := int64(mounted.Size)
	if err := binio.EnsureRange(file, partStart, partSize); err != nil {
		return err
	}

	layout, err := CalculateLayout(partStart, partSize)
	if err != nil {
		return err
	}

	sizeBefore, err := binio.FileSize(file)
	if err != nil {
		return err
	}

	if err := clearRange(file, partStart, partSize); err != nil {
		return err
	}

	sb, err := buildSuperBlock(layout)
	if err != nil {
		return err
	}
	if err := initializeEXT2(file, sb, layout); err != nil {
		return err
	}

	sizeAfter, err := binio.FileSize(file)
	if err != nil {
		return err
	}
	if sizeAfter != sizeBefore {
		return fmt.Errorf("el disco cambio de tamaño inesperadamente durante mkfs")
	}

	fmt.Fprintf(out, "Particion formateada EXT2: %s\n", opts.ID)
	return nil
}

func buildSuperBlock(layout Layout) (structs.SuperBlock, error) {
	inodeSize, err := binio.BinarySize(structs.Inode{})
	if err != nil {
		return structs.SuperBlock{}, err
	}

	now := structs.NowDateBytes()
	return structs.SuperBlock{
		SFilesystemType:  2,
		SInodesCount:     layout.InodeCount,
		SBlocksCount:     layout.BlockCount,
		SFreeBlocksCount: layout.BlockCount - 2,
		SFreeInodesCount: layout.InodeCount - 2,
		SMtime:           now,
		SUmTime:          now,
		SMntCount:        1,
		SMagic:           0xEF53,
		SInodeSize:       int32(inodeSize),
		SBlockSize:       BlockSize,
		// SFirstIno/SFirstBlo are free indexes; base offsets live in SInodeStart/SBlockStart.
		SFirstIno:     2,
		SFirstBlo:     2,
		SBmInodeStart: int32(layout.BmInodeStart),
		SBmBlockStart: int32(layout.BmBlockStart),
		SInodeStart:   int32(layout.InodeStart),
		SBlockStart:   int32(layout.BlockStart),
	}, nil
}

func initializeEXT2(file *os.File, sb structs.SuperBlock, layout Layout) error {
	inodeBitmap := make([]byte, layout.InodeCount)
	blockBitmap := make([]byte, layout.BlockCount)
	for _, index := range []int32{0, 1} {
		if err := MarkBitmapUsed(inodeBitmap, index); err != nil {
			return err
		}
		if err := MarkBitmapUsed(blockBitmap, index); err != nil {
			return err
		}
	}

	if err := WriteSuperBlock(file, layout.SuperStart, sb); err != nil {
		return err
	}
	if err := WriteBitmap(file, int64(sb.SBmInodeStart), inodeBitmap); err != nil {
		return err
	}
	if err := WriteBitmap(file, int64(sb.SBmBlockStart), blockBitmap); err != nil {
		return err
	}
	if err := initializeInodeTable(file, sb); err != nil {
		return err
	}

	rootInode := structs.NewEmptyInode()
	rootInode.IUid = 1
	rootInode.IGid = 1
	rootInode.ISize = 0
	now := structs.NowDateBytes()
	rootInode.IAtime = now
	rootInode.ICtime = now
	rootInode.IMtime = now
	rootInode.IBlock[0] = 0
	rootInode.IType = '0'
	structs.SetPerm(&rootInode.IPerm, "777")

	usersInode := structs.NewEmptyInode()
	usersInode.IUid = 1
	usersInode.IGid = 1
	usersInode.ISize = int32(len(initialUsersContent))
	usersInode.IAtime = now
	usersInode.ICtime = now
	usersInode.IMtime = now
	usersInode.IBlock[0] = 1
	usersInode.IType = '1'
	structs.SetPerm(&usersInode.IPerm, "664")

	rootBlock := structs.NewEmptyFolderBlock()
	structs.SetName12(&rootBlock.BContent[0].BName, ".")
	rootBlock.BContent[0].BInodo = 0
	structs.SetName12(&rootBlock.BContent[1].BName, "..")
	rootBlock.BContent[1].BInodo = 0
	structs.SetName12(&rootBlock.BContent[2].BName, "users.txt")
	rootBlock.BContent[2].BInodo = 1

	usersBlock := structs.NewEmptyFileBlock()
	structs.CopyString(usersBlock.BContent[:], initialUsersContent)

	if err := WriteInode(file, sb, 0, rootInode); err != nil {
		return err
	}
	if err := WriteInode(file, sb, 1, usersInode); err != nil {
		return err
	}
	if err := WriteFolderBlock(file, sb, 0, rootBlock); err != nil {
		return err
	}
	if err := WriteFileBlock(file, sb, 1, usersBlock); err != nil {
		return err
	}
	return nil
}

func initializeInodeTable(file *os.File, sb structs.SuperBlock) error {
	empty := structs.NewEmptyInode()
	for i := int32(0); i < sb.SInodesCount; i++ {
		if err := WriteInode(file, sb, i, empty); err != nil {
			return err
		}
	}
	return nil
}

func clearRange(file *os.File, offset int64, size int64) error {
	const chunkSize int64 = 1024 * 1024
	zeroes := make([]byte, chunkSize)
	written := int64(0)
	for written < size {
		current := chunkSize
		remaining := size - written
		if remaining < current {
			current = remaining
		}
		if err := binio.WriteBytesAt(file, offset+written, zeroes[:int(current)]); err != nil {
			return err
		}
		written += current
	}
	return nil
}
