package reports

import (
	"sort"
	"strings"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/partition"
	"mia_p1_201800996/internal/structs"
)

type diskSegment struct {
	label string
	size  int64
}

func BuildDiskDot(diskPath string, mbr structs.MBR) (string, error) {
	total, err := diskSize(diskPath)
	if err != nil {
		return "", err
	}

	segments, err := diskSegments(diskPath, mbr, total)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("digraph Reporte_DISK {\n")
	b.WriteString("  graph [rankdir=LR];\n")
	b.WriteString("  node [shape=plain];\n")
	b.WriteString("  disk [label=<\n")
	b.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\"><TR>\n")
	for _, segment := range segments {
		b.WriteString("      ")
		b.WriteString(htmlCell(htmlLabel(segment.label), esc(pct(segment.size, total))))
		b.WriteString("\n")
	}
	b.WriteString("    </TR></TABLE>\n")
	b.WriteString("  >];\n")
	b.WriteString("}\n")
	return b.String(), nil
}

func diskSegments(diskPath string, mbr structs.MBR, total int64) ([]diskSegment, error) {
	mbrSize, err := binio.BinarySize(structs.MBR{})
	if err != nil {
		return nil, err
	}

	result := []diskSegment{{label: "MBR", size: mbrSize}}
	active := activePartitions(mbr)
	cursor := mbrSize

	for _, part := range active {
		start := int64(part.PartStart)
		if start > cursor {
			result = append(result, diskSegment{label: "Libre", size: start - cursor})
		}

		name := structs.FixedBytesToString(part.PartName[:])
		switch part.PartType {
		case 'E':
			label, err := extendedLabel(diskPath, part, total)
			if err != nil {
				return nil, err
			}
			result = append(result, diskSegment{label: label, size: int64(part.PartSize)})
		default:
			result = append(result, diskSegment{label: "Primaria<br/>" + name, size: int64(part.PartSize)})
		}
		cursor = start + int64(part.PartSize)
	}

	if total > cursor {
		result = append(result, diskSegment{label: "Libre", size: total - cursor})
	}
	return result, nil
}

func activePartitions(mbr structs.MBR) []structs.Partition {
	result := make([]structs.Partition, 0, len(mbr.MbrPartitions))
	for _, part := range mbr.MbrPartitions {
		if part.PartStart >= 0 && part.PartSize > 0 && part.PartStatus != '0' {
			result = append(result, part)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].PartStart < result[j].PartStart
	})
	return result
}

func extendedLabel(diskPath string, ext structs.Partition, total int64) (string, error) {
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	chain, err := partition.ReadEBRChain(file, ext)
	if err != nil {
		return "", err
	}

	name := structs.FixedBytesToString(ext.PartName[:])
	parts := []string{"Extendida<br/>" + name + "<br/>" + pct(int64(ext.PartSize), total)}
	extendedEnd := int64(ext.PartStart + ext.PartSize)
	cursor := int64(ext.PartStart)
	ebrSize, err := binio.BinarySize(structs.EBR{})
	if err != nil {
		return "", err
	}

	for _, item := range chain {
		if item.Offset > cursor {
			parts = append(parts, "Libre<br/>"+pct(item.Offset-cursor, total))
		}
		parts = append(parts, "EBR<br/>"+pct(ebrSize, total))
		cursor = item.Offset + ebrSize

		if item.EBR.PartSize > 0 {
			logicalStart := int64(item.EBR.PartStart)
			if logicalStart > cursor {
				parts = append(parts, "Libre<br/>"+pct(logicalStart-cursor, total))
			}
			logicalName := structs.FixedBytesToString(item.EBR.PartName[:])
			parts = append(parts, "Logica<br/>"+logicalName+"<br/>"+pct(int64(item.EBR.PartSize), total))
			cursor = logicalStart + int64(item.EBR.PartSize)
		}
	}

	if extendedEnd > cursor {
		parts = append(parts, "Libre<br/>"+pct(extendedEnd-cursor, total))
	}
	return strings.Join(parts, "<br/>"), nil
}
