package services

import (
	"fmt"
	"strings"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/partition"
	"mia_p1_201800996/internal/structs"
)

func ListPartitions(path string) ([]dto.PartitionResponse, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("path es obligatorio")
	}
	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var mbr structs.MBR
	if err := binio.ReadStructAt(file, 0, &mbr); err != nil {
		return nil, err
	}

	partitions := make([]dto.PartitionResponse, 0)
	for _, part := range mbr.MbrPartitions {
		if part.PartStart < 0 || part.PartSize <= 0 {
			continue
		}
		partitions = append(partitions, dto.PartitionResponse{
			Name:   structs.FixedBytesToString(part.PartName[:]),
			Type:   byteString(part.PartType),
			Fit:    byteString(part.PartFit),
			Start:  part.PartStart,
			Size:   part.PartSize,
			Status: byteString(part.PartStatus),
		})
	}

	if extended, _, ok := partition.FindExtendedPartition(mbr); ok {
		chain, err := partition.ReadEBRChain(file, extended)
		if err != nil {
			return nil, err
		}
		for _, item := range chain {
			ebr := item.EBR
			if ebr.PartSize <= 0 {
				continue
			}
			partitions = append(partitions, dto.PartitionResponse{
				Name:   structs.FixedBytesToString(ebr.PartName[:]),
				Type:   "L",
				Fit:    byteString(ebr.PartFit),
				Start:  ebr.PartStart,
				Size:   ebr.PartSize,
				Status: byteString(ebr.PartMount),
			})
		}
	}

	return partitions, nil
}

func CreatePartition(req dto.CreatePartitionRequest) error {
	return partition.Create(partition.CreateOptions{
		Size: req.Size,
		Unit: req.Unit,
		Path: req.Path,
		Type: req.Type,
		Fit:  req.Fit,
		Name: req.Name,
	})
}

func byteString(value byte) string {
	if value == 0 {
		return ""
	}
	return string([]byte{value})
}
