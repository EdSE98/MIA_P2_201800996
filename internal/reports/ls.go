package reports

import (
	"fmt"
	"os"
	"strings"

	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/structs"
)

func BuildLSDot(file *os.File, sb structs.SuperBlock, path string) (string, error) {
	index, inode, err := fs.ResolvePath(file, sb, path)
	if err != nil {
		return "", err
	}
	rows, err := lsRows(file, sb, index, inode, path)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("digraph Reporte_LS {\n")
	b.WriteString("  node [shape=plain];\n")
	b.WriteString("  ls [label=<\n")
	b.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")
	b.WriteString("      <TR><TD><B>Permisos</B></TD><TD><B>UID</B></TD><TD><B>GID</B></TD><TD><B>Size</B></TD><TD><B>Modificacion</B></TD><TD><B>Tipo</B></TD><TD><B>Nombre</B></TD></TR>\n")
	for _, rowData := range rows {
		b.WriteString(fmt.Sprintf("      <TR><TD>%s</TD><TD>%d</TD><TD>%d</TD><TD>%d</TD><TD>%s</TD><TD>%s</TD><TD>%s</TD></TR>\n",
			esc(rowData.perm), rowData.uid, rowData.gid, rowData.size, esc(rowData.mtime), esc(rowData.kind), esc(rowData.name)))
	}
	b.WriteString("    </TABLE>\n")
	b.WriteString("  >];\n")
	b.WriteString("}\n")
	return b.String(), nil
}

type lsRow struct {
	perm  string
	uid   int32
	gid   int32
	size  int32
	mtime string
	kind  string
	name  string
}

func lsRows(file *os.File, sb structs.SuperBlock, index int32, inode structs.Inode, name string) ([]lsRow, error) {
	if inode.IType == '1' {
		return []lsRow{inodeToLSRow(inode, name)}, nil
	}
	var rows []lsRow
	for i := 0; i < 12 && i < len(inode.IBlock); i++ {
		blockIndex := inode.IBlock[i]
		if blockIndex < 0 {
			continue
		}
		block, err := fs.ReadFolderBlock(file, sb, blockIndex)
		if err != nil {
			return nil, err
		}
		for _, content := range block.BContent {
			entryName := structs.FixedBytesToString(content.BName[:])
			if content.BInodo < 0 || entryName == "." || entryName == ".." || entryName == "" {
				continue
			}
			child, err := fs.ReadInode(file, sb, content.BInodo)
			if err != nil {
				return nil, err
			}
			rows = append(rows, inodeToLSRow(child, entryName))
		}
	}
	_ = index
	return rows, nil
}

func inodeToLSRow(inode structs.Inode, name string) lsRow {
	kind := "Archivo"
	if inode.IType == '0' {
		kind = "Carpeta"
	}
	return lsRow{
		perm:  structs.FixedBytesToString(inode.IPerm[:]),
		uid:   inode.IUid,
		gid:   inode.IGid,
		size:  inode.ISize,
		mtime: structs.FixedBytesToString(inode.IMtime[:]),
		kind:  kind,
		name:  name,
	}
}
