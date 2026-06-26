package services

import (
	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/mount"
)

func ListMounts() []dto.MountResponse {
	mounted := mount.Global.List()
	result := make([]dto.MountResponse, 0, len(mounted))
	for _, item := range mounted {
		result = append(result, mountToDTO(item))
	}
	return result
}

func mountToDTO(item mount.MountedPartition) dto.MountResponse {
	return dto.MountResponse{
		ID:            item.ID,
		Path:          item.DiskPath,
		Name:          item.PartitionName,
		PartitionType: string([]byte{item.PartitionType}),
		Start:         item.Start,
		Size:          item.Size,
	}
}
