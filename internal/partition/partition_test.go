package partition

import (
	"path/filepath"
	"testing"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

func TestCreatePrimaryPartition(t *testing.T) {
	path := makeDisk(t, "primary.mia")

	if err := Create(CreateOptions{
		Size: 100,
		Unit: "K",
		Path: path,
		Type: "P",
		Fit:  "FF",
		Name: "Part1",
	}); err != nil {
		t.Fatalf("Create primary failed: %v", err)
	}

	mbr, err := disk.ReadMBR(path)
	if err != nil {
		t.Fatalf("ReadMBR failed: %v", err)
	}
	part := mbr.MbrPartitions[0]
	if structs.FixedBytesToString(part.PartName[:]) != "Part1" {
		t.Fatalf("name = %q, want Part1", structs.FixedBytesToString(part.PartName[:]))
	}
	if part.PartType != 'P' {
		t.Fatalf("type = %q, want P", part.PartType)
	}
	if part.PartStart != int32(disk.SizeOfMBR()) {
		t.Fatalf("start = %d, want %d", part.PartStart, disk.SizeOfMBR())
	}
	if part.PartSize != 100*1024 {
		t.Fatalf("size = %d, want %d", part.PartSize, 100*1024)
	}
	assertDiskSize(t, path, 1024*1024)
}

func TestCreateExtendedAndRejectSecondExtended(t *testing.T) {
	path := makeDisk(t, "extended.mia")

	if err := Create(CreateOptions{Size: 100, Unit: "K", Path: path, Type: "E", Name: "Ext1"}); err != nil {
		t.Fatalf("Create extended failed: %v", err)
	}
	if err := Create(CreateOptions{Size: 100, Unit: "K", Path: path, Type: "E", Name: "Ext2"}); err == nil {
		t.Fatal("expected second extended error")
	}

	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()

	mbr, err := disk.ReadMBR(path)
	if err != nil {
		t.Fatalf("ReadMBR failed: %v", err)
	}
	ext, _, ok := FindExtendedPartition(mbr)
	if !ok {
		t.Fatal("expected extended partition")
	}
	var ebr structs.EBR
	if err := binio.ReadStructAt(file, int64(ext.PartStart), &ebr); err != nil {
		t.Fatalf("Read EBR failed: %v", err)
	}
	if ebr.PartNext != -1 {
		t.Fatalf("empty EBR PartNext = %d, want -1", ebr.PartNext)
	}
	assertDiskSize(t, path, 1024*1024)
}

func TestCreatePartitionUsingWholeDiskLeavesMBRSpace(t *testing.T) {
	path := makeDisk(t, "whole-disk.mia")
	if err := Create(CreateOptions{Size: 1, Unit: "M", Path: path, Type: "E", Name: "Ext1"}); err != nil {
		t.Fatalf("Create whole disk extended failed: %v", err)
	}
	mbr, err := disk.ReadMBR(path)
	if err != nil {
		t.Fatalf("ReadMBR failed: %v", err)
	}
	part := mbr.MbrPartitions[0]
	if part.PartStart != int32(disk.SizeOfMBR()) {
		t.Fatalf("start = %d, want %d", part.PartStart, disk.SizeOfMBR())
	}
	wantSize := int32(1024*1024 - disk.SizeOfMBR())
	if part.PartSize != wantSize {
		t.Fatalf("size = %d, want %d", part.PartSize, wantSize)
	}
	assertDiskSize(t, path, 1024*1024)
}

func TestCreateLogicalWithoutExtendedFails(t *testing.T) {
	path := makeDisk(t, "logical-no-ext.mia")
	err := Create(CreateOptions{Size: 100, Unit: "K", Path: path, Type: "L", Name: "Log1"})
	if err == nil {
		t.Fatal("expected logical without extended error")
	}
}

func TestCreateLogicalPartition(t *testing.T) {
	path := makeDisk(t, "logical.mia")
	if err := Create(CreateOptions{Size: 500, Unit: "K", Path: path, Type: "E", Name: "Ext1"}); err != nil {
		t.Fatalf("Create extended failed: %v", err)
	}
	if err := Create(CreateOptions{Size: 100, Unit: "K", Path: path, Type: "L", Name: "Log1"}); err != nil {
		t.Fatalf("Create logical failed: %v", err)
	}

	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()

	mbr, err := disk.ReadMBR(path)
	if err != nil {
		t.Fatalf("ReadMBR failed: %v", err)
	}
	ext, _, ok := FindExtendedPartition(mbr)
	if !ok {
		t.Fatal("expected extended partition")
	}
	chain, err := ReadEBRChain(file, ext)
	if err != nil {
		t.Fatalf("ReadEBRChain failed: %v", err)
	}
	if len(chain) != 1 {
		t.Fatalf("chain length = %d, want 1", len(chain))
	}
	if structs.FixedBytesToString(chain[0].EBR.PartName[:]) != "Log1" {
		t.Fatalf("logical name = %q, want Log1", structs.FixedBytesToString(chain[0].EBR.PartName[:]))
	}
}

func TestCreateFifthPrimaryOrExtendedFails(t *testing.T) {
	path := makeDisk(t, "fifth.mia")
	for i := 1; i <= 4; i++ {
		if err := Create(CreateOptions{Size: 100, Unit: "K", Path: path, Type: "P", Name: "Part" + string(byte('0'+i))}); err != nil {
			t.Fatalf("Create partition %d failed: %v", i, err)
		}
	}
	if err := Create(CreateOptions{Size: 100, Unit: "K", Path: path, Type: "P", Name: "Part5"}); err == nil {
		t.Fatal("expected fifth partition error")
	}
}

func TestCreateDuplicateNameFails(t *testing.T) {
	path := makeDisk(t, "duplicate.mia")
	if err := Create(CreateOptions{Size: 100, Unit: "K", Path: path, Type: "P", Name: "Part1"}); err != nil {
		t.Fatalf("Create partition failed: %v", err)
	}
	if err := Create(CreateOptions{Size: 100, Unit: "K", Path: path, Type: "P", Name: "Part1"}); err == nil {
		t.Fatal("expected duplicate name error")
	}
}

func makeDisk(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := disk.MakeDisk(disk.MakeDiskOptions{Size: 1, Unit: "M", Path: path}); err != nil {
		t.Fatalf("MakeDisk failed: %v", err)
	}
	return path
}

func assertDiskSize(t *testing.T, path string, want int64) {
	t.Helper()
	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()
	size, err := binio.FileSize(file)
	if err != nil {
		t.Fatalf("FileSize failed: %v", err)
	}
	if size != want {
		t.Fatalf("disk size = %d, want %d", size, want)
	}
}
