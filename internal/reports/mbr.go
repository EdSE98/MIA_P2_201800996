package reports

import (
	"fmt"
	"os"
	"strings"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/partition"
	"mia_p1_201800996/internal/structs"
)

func BuildMBRDot(diskPath string, mbr structs.MBR) (string, error) {
	var b strings.Builder
	b.WriteString("digraph Reporte_MBR {\n")
	b.WriteString("  graph [rankdir=TB];\n")
	b.WriteString("  node [shape=plain];\n")
	b.WriteString("  mbr [label=<\n")
	b.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")
	b.WriteString("      <TR><TD COLSPAN=\"2\"><B>Reporte_MBR</B></TD></TR>\n")
	row(&b, "mbr_tamano", fmt.Sprintf("%d", mbr.MbrTamano))
	row(&b, "mbr_fecha_creacion", structs.FixedBytesToString(mbr.MbrFechaCreacion[:]))
	row(&b, "mbr_dsk_signature", fmt.Sprintf("%d", mbr.MbrDskSignature))
	row(&b, "dsk_fit", byteText(mbr.DskFit))

	for i, part := range mbr.MbrPartitions {
		b.WriteString(fmt.Sprintf("      <TR><TD COLSPAN=\"2\"><B>Partition_%d</B></TD></TR>\n", i+1))
		row(&b, "part_status", byteText(part.PartStatus))
		row(&b, "part_type", byteText(part.PartType))
		row(&b, "part_fit", byteText(part.PartFit))
		row(&b, "part_start", fmt.Sprintf("%d", part.PartStart))
		row(&b, "part_s", fmt.Sprintf("%d", part.PartSize))
		row(&b, "part_name", structs.FixedBytesToString(part.PartName[:]))
		row(&b, "part_correlative", fmt.Sprintf("%d", part.PartCorrelative))
		row(&b, "part_id", structs.FixedBytesToString(part.PartID[:]))
	}

	if ext, _, ok := partition.FindExtendedPartition(mbr); ok {
		file, _, err := disk.OpenReadWrite(diskPath)
		if err != nil {
			return "", err
		}
		defer file.Close()
		chain, err := partition.ReadEBRChain(file, ext)
		if err != nil {
			return "", err
		}
		for i, item := range chain {
			b.WriteString(fmt.Sprintf("      <TR><TD COLSPAN=\"2\"><B>EBR_%d</B></TD></TR>\n", i+1))
			row(&b, "part_mount", byteText(item.EBR.PartMount))
			row(&b, "part_fit", byteText(item.EBR.PartFit))
			row(&b, "part_start", fmt.Sprintf("%d", item.EBR.PartStart))
			row(&b, "part_s", fmt.Sprintf("%d", item.EBR.PartSize))
			row(&b, "part_next", fmt.Sprintf("%d", item.EBR.PartNext))
			row(&b, "part_name", structs.FixedBytesToString(item.EBR.PartName[:]))
		}
	}

	b.WriteString("    </TABLE>\n")
	b.WriteString("  >];\n")
	b.WriteString("}\n")
	return b.String(), nil
}

func row(b *strings.Builder, key string, value string) {
	b.WriteString(fmt.Sprintf("      <TR><TD><B>%s</B></TD><TD>%s</TD></TR>\n", esc(key), esc(value)))
}

func diskSize(path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
