package services

import (
	"path/filepath"
	"testing"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/mount"
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

func TestDeletePartitionRejectsMountedPartition(t *testing.T) {
	oldMount := mount.Global
	mount.Global = mount.NewManager()
	t.Cleanup(func() { mount.Global = oldMount })

	path := filepath.Join(t.TempDir(), "mounted_partition.dsk")
	if _, err := CreateDisk(dto.CreateDiskRequest{Size: 2, Unit: "M", Path: path}); err != nil {
		t.Fatal(err)
	}
	if err := CreatePartition(dto.CreatePartitionRequest{
		Path: path, Name: "Part1", Size: 512, Unit: "K", Type: "P",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := mount.Global.Mount(path, "Part1"); err != nil {
		t.Fatal(err)
	}

	err := DeletePartition(dto.DeletePartitionRequest{Path: path, Name: "Part1", Delete: "fast"})
	if err == nil {
		t.Fatal("expected mounted partition delete error")
	}
}

func TestResizeAndDeletePartitionServices(t *testing.T) {
	oldMount := mount.Global
	mount.Global = mount.NewManager()
	t.Cleanup(func() { mount.Global = oldMount })

	path := filepath.Join(t.TempDir(), "partition_operations.dsk")
	if _, err := CreateDisk(dto.CreateDiskRequest{Size: 3, Unit: "M", Path: path}); err != nil {
		t.Fatal(err)
	}
	if err := CreatePartition(dto.CreatePartitionRequest{
		Path: path, Name: "Part1", Size: 512, Unit: "K", Type: "P",
	}); err != nil {
		t.Fatal(err)
	}
	if err := ResizePartition(dto.ResizePartitionRequest{
		Path: path, Name: "Part1", Add: 128, Unit: "K",
	}); err != nil {
		t.Fatal(err)
	}
	partitions, err := ListPartitions(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(partitions) != 1 || partitions[0].Size != 640*1024 {
		t.Fatalf("unexpected resized partition: %+v", partitions)
	}
	if err := DeletePartition(dto.DeletePartitionRequest{
		Path: path, Name: "Part1", Delete: "fast",
	}); err != nil {
		t.Fatal(err)
	}
	partitions, err = ListPartitions(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(partitions) != 0 {
		t.Fatalf("expected no partitions, got %+v", partitions)
	}
}
