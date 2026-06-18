package reports

import (
	"fmt"
	"os"
	"strings"

	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/structs"
)

func BuildInodeDot(file *os.File, sb structs.SuperBlock) (string, error) {
	indices, err := UsedInodeIndices(file, sb)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("digraph Reporte_Inode {\n")
	b.WriteString("  node [shape=plain];\n")
	for _, index := range indices {
		inode, err := fs.ReadInode(file, sb, index)
		if err != nil {
			return "", err
		}
		b.WriteString(fmt.Sprintf("  inode_%d [label=<\n", index))
		b.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")
		b.WriteString(fmt.Sprintf("      <TR><TD COLSPAN=\"2\"><B>Inodo %d</B></TD></TR>\n", index))
		row(&b, "i_uid", fmt.Sprintf("%d", inode.IUid))
		row(&b, "i_gid", fmt.Sprintf("%d", inode.IGid))
		row(&b, "i_s", fmt.Sprintf("%d", inode.ISize))
		row(&b, "i_atime", structs.FixedBytesToString(inode.IAtime[:]))
		row(&b, "i_ctime", structs.FixedBytesToString(inode.ICtime[:]))
		row(&b, "i_mtime", structs.FixedBytesToString(inode.IMtime[:]))
		for i, ptr := range inode.IBlock {
			row(&b, fmt.Sprintf("i_block[%d]", i), fmt.Sprintf("%d", ptr))
		}
		row(&b, "i_type", byteText(inode.IType))
		row(&b, "i_perm", structs.FixedBytesToString(inode.IPerm[:]))
		b.WriteString("    </TABLE>\n")
		b.WriteString("  >];\n")
	}
	b.WriteString("}\n")
	return b.String(), nil
}
