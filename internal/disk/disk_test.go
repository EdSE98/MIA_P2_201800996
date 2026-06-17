package disk

import (
	"path/filepath"
	"testing"

	"mia_p1_201800996/internal/structs"
)

func TestMakeDiskCreatesFixedSizeDiskAndMBR(t *testing.T) {
	path := filepath.Join(t.TempDir(), "disk1.mia")

	if err := MakeDisk(MakeDiskOptions{
		Size: 1,
		Unit: "M",
		Fit:  "BF",
		Path: path,
	}); err != nil {
		t.Fatalf("MakeDisk failed: %v", err)
	}

	mbr, err := ReadMBR(path)
	if err != nil {
		t.Fatalf("ReadMBR failed: %v", err)
	}
	if mbr.MbrTamano != 1024*1024 {
		t.Fatalf("MbrTamano = %d, want %d", mbr.MbrTamano, 1024*1024)
	}
	if mbr.DskFit != 'B' {
		t.Fatalf("DskFit = %q, want B", mbr.DskFit)
	}
	for i, part := range mbr.MbrPartitions {
		if part.PartStart != -1 || part.PartSize != 0 || part.PartStatus != '0' || part.PartCorrelative != 0 {
			t.Fatalf("partition %d not empty: %#v", i, part)
		}
	}

	file, _, err := OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()
	size, err := file.Stat()
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if size.Size() != 1024*1024 {
		t.Fatalf("file size = %d, want %d", size.Size(), 1024*1024)
	}
}

func TestMakeDiskRejectsInvalidSize(t *testing.T) {
	err := MakeDisk(MakeDiskOptions{
		Size: 0,
		Path: filepath.Join(t.TempDir(), "disk.mia"),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMakeDiskRejectsExistingPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "disk.mia")
	if err := MakeDisk(MakeDiskOptions{Size: 1, Unit: "M", Path: path}); err != nil {
		t.Fatalf("MakeDisk failed: %v", err)
	}
	if err := MakeDisk(MakeDiskOptions{Size: 1, Unit: "M", Path: path}); err == nil {
		t.Fatal("expected existing path error")
	}
}

func TestReadWriteMBR(t *testing.T) {
	path := filepath.Join(t.TempDir(), "disk.mia")
	if err := MakeDisk(MakeDiskOptions{Size: 1, Unit: "M", Path: path}); err != nil {
		t.Fatalf("MakeDisk failed: %v", err)
	}

	mbr, err := ReadMBR(path)
	if err != nil {
		t.Fatalf("ReadMBR failed: %v", err)
	}
	structs.SetName16(&mbr.MbrPartitions[0].PartName, "Part1")
	mbr.MbrPartitions[0].PartStatus = '1'
	if err := WriteMBR(path, mbr); err != nil {
		t.Fatalf("WriteMBR failed: %v", err)
	}

	got, err := ReadMBR(path)
	if err != nil {
		t.Fatalf("ReadMBR failed: %v", err)
	}
	if structs.FixedBytesToString(got.MbrPartitions[0].PartName[:]) != "Part1" {
		t.Fatalf("partition name not persisted")
	}
}
