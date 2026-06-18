package reports

import (
	"fmt"
	"os"
	"strings"

	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/structs"
)

func BuildTreeDot(file *os.File, sb structs.SuperBlock) (string, error) {
	var b strings.Builder
	b.WriteString("digraph Reporte_Tree {\n")
	b.WriteString("  node [shape=plain];\n")
	visited := map[int32]bool{}
	if err := writeTreeInode(&b, file, sb, 0, visited); err != nil {
		return "", err
	}
	b.WriteString("}\n")
	return b.String(), nil
}

func writeTreeInode(b *strings.Builder, file *os.File, sb structs.SuperBlock, inodeIndex int32, visited map[int32]bool) error {
	if visited[inodeIndex] {
		return nil
	}
	visited[inodeIndex] = true
	inode, err := fs.ReadInode(file, sb, inodeIndex)
	if err != nil {
		return err
	}
	b.WriteString(fmt.Sprintf("  inode_%d [label=<\n", inodeIndex))
	b.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")
	b.WriteString(fmt.Sprintf("      <TR><TD COLSPAN=\"2\"><B>Inodo %d</B></TD></TR>\n", inodeIndex))
	row(b, "uid", fmt.Sprintf("%d", inode.IUid))
	row(b, "gid", fmt.Sprintf("%d", inode.IGid))
	row(b, "size", fmt.Sprintf("%d", inode.ISize))
	row(b, "type", byteText(inode.IType))
	row(b, "perm", structs.FixedBytesToString(inode.IPerm[:]))
	for i, ptr := range inode.IBlock {
		row(b, fmt.Sprintf("block[%d]", i), fmt.Sprintf("%d", ptr))
	}
	b.WriteString("    </TABLE>\n")
	b.WriteString("  >];\n")

	for i := 0; i < 12 && i < len(inode.IBlock); i++ {
		blockIndex := inode.IBlock[i]
		if blockIndex < 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("  inode_%d -> block_%d;\n", inodeIndex, blockIndex))
		if inode.IType == '0' {
			if err := writeTreeFolderBlock(b, file, sb, blockIndex, visited); err != nil {
				return err
			}
		} else {
			if err := writeTreeFileBlock(b, file, sb, blockIndex); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeTreeFolderBlock(b *strings.Builder, file *os.File, sb structs.SuperBlock, blockIndex int32, visited map[int32]bool) error {
	block, err := fs.ReadFolderBlock(file, sb, blockIndex)
	if err != nil {
		return err
	}
	b.WriteString(fmt.Sprintf("  block_%d [label=<\n", blockIndex))
	b.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")
	b.WriteString(fmt.Sprintf("      <TR><TD COLSPAN=\"2\"><B>Bloque Carpeta %d</B></TD></TR>\n", blockIndex))
	for _, content := range block.BContent {
		name := structs.FixedBytesToString(content.BName[:])
		row(b, name, fmt.Sprintf("%d", content.BInodo))
	}
	b.WriteString("    </TABLE>\n")
	b.WriteString("  >];\n")
	for _, content := range block.BContent {
		name := structs.FixedBytesToString(content.BName[:])
		if content.BInodo < 0 || name == "." || name == ".." || name == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("  block_%d -> inode_%d [label=\"%s\"];\n", blockIndex, content.BInodo, esc(name)))
		if err := writeTreeInode(b, file, sb, content.BInodo, visited); err != nil {
			return err
		}
	}
	return nil
}

func writeTreeFileBlock(b *strings.Builder, file *os.File, sb structs.SuperBlock, blockIndex int32) error {
	block, err := fs.ReadFileBlock(file, sb, blockIndex)
	if err != nil {
		return err
	}
	b.WriteString(fmt.Sprintf("  block_%d [label=<\n", blockIndex))
	b.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")
	b.WriteString(fmt.Sprintf("      <TR><TD><B>Bloque Archivo %d</B></TD></TR>\n", blockIndex))
	b.WriteString(fmt.Sprintf("      <TR><TD>%s</TD></TR>\n", esc(cleanBlockText(block.BContent[:]))))
	b.WriteString("    </TABLE>\n")
	b.WriteString("  >];\n")
	return nil
}
