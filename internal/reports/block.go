package reports

import (
	"fmt"
	"os"
	"strings"

	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/structs"
)

func BuildBlockDot(file *os.File, sb structs.SuperBlock) (string, error) {
	blockTypes, err := referencedBlockTypes(file, sb)
	if err != nil {
		return "", err
	}
	indices, err := UsedBlockIndices(file, sb)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("digraph Reporte_Block {\n")
	b.WriteString("  node [shape=plain];\n")
	for _, index := range indices {
		switch blockTypes[index] {
		case '0':
			block, err := fs.ReadFolderBlock(file, sb, index)
			if err != nil {
				return "", err
			}
			b.WriteString(fmt.Sprintf("  block_%d [label=<\n", index))
			b.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")
			b.WriteString(fmt.Sprintf("      <TR><TD COLSPAN=\"2\"><B>Bloque Carpeta %d</B></TD></TR>\n", index))
			for _, content := range block.BContent {
				row(&b, "b_name", structs.FixedBytesToString(content.BName[:]))
				row(&b, "b_inodo", fmt.Sprintf("%d", content.BInodo))
			}
			b.WriteString("    </TABLE>\n")
			b.WriteString("  >];\n")
		default:
			block, err := fs.ReadFileBlock(file, sb, index)
			if err != nil {
				return "", err
			}
			b.WriteString(fmt.Sprintf("  block_%d [label=<\n", index))
			b.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")
			b.WriteString(fmt.Sprintf("      <TR><TD><B>Bloque Archivo %d</B></TD></TR>\n", index))
			b.WriteString(fmt.Sprintf("      <TR><TD>%s</TD></TR>\n", esc(cleanBlockText(block.BContent[:]))))
			b.WriteString("    </TABLE>\n")
			b.WriteString("  >];\n")
		}
	}
	b.WriteString("}\n")
	return b.String(), nil
}

func referencedBlockTypes(file *os.File, sb structs.SuperBlock) (map[int32]byte, error) {
	result := map[int32]byte{}
	inodes, err := UsedInodeIndices(file, sb)
	if err != nil {
		return nil, err
	}
	for _, index := range inodes {
		inode, err := fs.ReadInode(file, sb, index)
		if err != nil {
			return nil, err
		}
		for i := 0; i < 12 && i < len(inode.IBlock); i++ {
			if inode.IBlock[i] >= 0 {
				result[inode.IBlock[i]] = inode.IType
			}
		}
	}
	return result, nil
}
