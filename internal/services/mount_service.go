package services

import (
	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/mount"
)

func ListMounts() []dto.MountResponse {
	mounted := mount.Global.List()
	result := make([]dto.MountResponse, 0, len(mounted))
	for _, item := range mounted {
		result = append(result, dto.MountResponse{
			ID:            item.ID,
			DiskPath:      item.DiskPath,
			PartitionName: item.PartitionName,
			PartitionType: string([]byte{item.PartitionType}),
			Start:         item.Start,
			Size:          item.Size,
		})
	}
	return result
}
