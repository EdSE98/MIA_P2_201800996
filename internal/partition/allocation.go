package partition

import (
	"fmt"
	"sort"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

type freeSpace struct {
	start int64
	size  int64
}

type usedPartition struct {
	start int64
	size  int64
}

func allocatePrimaryOrExtended(mbr structs.MBR, size int64, fit byte) (int64, error) {
	spaces := freeSpaces(mbr)
	var selected *freeSpace

	for i := range spaces {
		space := &spaces[i]
		if space.size < size {
			continue
		}
		switch fit {
		case 'F':
			return space.start, nil
		case 'B':
			if selected == nil || space.size < selected.size {
				selected = space
			}
		case 'W':
			if selected == nil || space.size > selected.size {
				selected = space
			}
		default:
			return 0, fmt.Errorf("fit invalido %q", fit)
		}
	}

	if selected == nil {
		return 0, fmt.Errorf("no hay espacio suficiente para la particion")
	}
	return selected.start, nil
}

func freeSpaces(mbr structs.MBR) []freeSpace {
	used := make([]usedPartition, 0, len(mbr.MbrPartitions))
	for _, part := range mbr.MbrPartitions {
		if isActivePartition(part) {
			used = append(used, usedPartition{
				start: int64(part.PartStart),
				size:  int64(part.PartSize),
			})
		}
	}
	sort.Slice(used, func(i, j int) bool {
		return used[i].start < used[j].start
	})

	spaces := make([]freeSpace, 0, len(used)+1)
	cursor := disk.SizeOfMBR()
	for _, part := range used {
		if part.start > cursor {
			spaces = append(spaces, freeSpace{
				start: cursor,
				size:  part.start - cursor,
			})
		}
		end := part.start + part.size
		if end > cursor {
			cursor = end
		}
	}

	diskEnd := int64(mbr.MbrTamano)
	if diskEnd > cursor {
		spaces = append(spaces, freeSpace{
			start: cursor,
			size:  diskEnd - cursor,
		})
	}
	return spaces
}

func isActivePartition(part structs.Partition) bool {
	return part.PartStart >= 0 && part.PartSize > 0 && part.PartStatus != '0'
}
