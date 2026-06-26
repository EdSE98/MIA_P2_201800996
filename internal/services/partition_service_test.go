package services

import (
	"path/filepath"
	"testing"

	"mia_p1_201800996/internal/api/dto"
)

func TestCreateAndListPartitions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "api_partition.dsk")
	if _, err := CreateDisk(dto.CreateDiskRequest{
		Size: 2,
		Unit: "M",
		Fit:  "FF",
		Path: path,
	}); err != nil {
		t.Fatal(err)
	}

	if err := CreatePartition(dto.CreatePartitionRequest{
		Path: path,
		Name: "Part1",
		Size: 512,
		Unit: "K",
		Type: "P",
		Fit:  "FF",
	}); err != nil {
		t.Fatal(err)
	}

	partitions, err := ListPartitions(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(partitions) != 1 {
		t.Fatalf("expected 1 partition, got %d", len(partitions))
	}
	if partitions[0].Name != "Part1" || partitions[0].Type != "P" || partitions[0].Size != 512*1024 {
		t.Fatalf("unexpected partition: %+v", partitions[0])
	}
}
