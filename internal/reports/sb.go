package reports

import (
	"fmt"
	"strings"

	"mia_p1_201800996/internal/structs"
)

func BuildSuperBlockDot(sb structs.SuperBlock) string {
	var b strings.Builder
	b.WriteString("digraph Reporte_SB {\n")
	b.WriteString("  node [shape=plain];\n")
	b.WriteString("  sb [label=<\n")
	b.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")
	b.WriteString("      <TR><TD COLSPAN=\"2\"><B>Reporte SuperBlock</B></TD></TR>\n")
	row(&b, "s_filesystem_type", fmt.Sprintf("%d", sb.SFilesystemType))
	row(&b, "s_inodes_count", fmt.Sprintf("%d", sb.SInodesCount))
	row(&b, "s_blocks_count", fmt.Sprintf("%d", sb.SBlocksCount))
	row(&b, "s_free_blocks_count", fmt.Sprintf("%d", sb.SFreeBlocksCount))
	row(&b, "s_free_inodes_count", fmt.Sprintf("%d", sb.SFreeInodesCount))
	row(&b, "s_mtime", structs.FixedBytesToString(sb.SMtime[:]))
	row(&b, "s_umtime", structs.FixedBytesToString(sb.SUmTime[:]))
	row(&b, "s_mnt_count", fmt.Sprintf("%d", sb.SMntCount))
	row(&b, "s_magic", fmt.Sprintf("%d (0x%X)", sb.SMagic, sb.SMagic))
	row(&b, "s_inode_s", fmt.Sprintf("%d", sb.SInodeSize))
	row(&b, "s_block_s", fmt.Sprintf("%d", sb.SBlockSize))
	row(&b, "s_first_ino", fmt.Sprintf("%d", sb.SFirstIno))
	row(&b, "s_first_blo", fmt.Sprintf("%d", sb.SFirstBlo))
	row(&b, "s_bm_inode_start", fmt.Sprintf("%d", sb.SBmInodeStart))
	row(&b, "s_bm_block_start", fmt.Sprintf("%d", sb.SBmBlockStart))
	row(&b, "s_inode_start", fmt.Sprintf("%d", sb.SInodeStart))
	row(&b, "s_block_start", fmt.Sprintf("%d", sb.SBlockStart))
	b.WriteString("    </TABLE>\n")
	b.WriteString("  >];\n")
	b.WriteString("}\n")
	return b.String()
}
