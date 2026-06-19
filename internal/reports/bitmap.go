package reports

import (
	"fmt"
	"strings"
)

func BuildBitmapText(bitmap []byte) string {
	var b strings.Builder
	for i, value := range bitmap {
		if i > 0 {
			if i%20 == 0 {
				b.WriteByte('\n')
			} else {
				b.WriteByte(' ')
			}
		}
		if value == 1 {
			b.WriteByte('1')
		} else {
			b.WriteByte('0')
		}
	}
	b.WriteByte('\n')
	return b.String()
}

func BuildBitmapDot(title string, bitmap []byte) string {
	lines := strings.Split(strings.TrimSpace(BuildBitmapText(bitmap)), "\n")
	var b strings.Builder
	b.WriteString("digraph Reporte_Bitmap {\n")
	b.WriteString("  node [shape=plain];\n")
	b.WriteString("  bitmap [label=<\n")
	b.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")
	b.WriteString(fmt.Sprintf("      <TR><TD><B>%s</B></TD></TR>\n", esc(title)))
	for _, line := range lines {
		b.WriteString(fmt.Sprintf("      <TR><TD>%s</TD></TR>\n", esc(line)))
	}
	b.WriteString("    </TABLE>\n")
	b.WriteString("  >];\n")
	b.WriteString("}\n")
	return b.String()
}
