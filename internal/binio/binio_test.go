package binio

import (
	"os"
	"path/filepath"
	"testing"

	"mia_p1_201800996/internal/structs"
)

func TestStructReadWriteDoesNotGrowFile(t *testing.T) {
	file := createFixedTempFile(t, 1024)
	defer file.Close()

	mbr := structs.NewEmptyMBR()
	mbr.MbrTamano = 1024
	mbr.MbrDskSignature = 201800996
	mbr.DskFit = 'F'
	structs.SetName16(&mbr.MbrPartitions[0].PartName, "Part1")

	if err := WriteStructAt(file, 0, mbr); err != nil {
		t.Fatalf("WriteStructAt failed: %v", err)
	}

	var got structs.MBR
	if err := ReadStructAt(file, 0, &got); err != nil {
		t.Fatalf("ReadStructAt failed: %v", err)
	}

	if got.MbrTamano != mbr.MbrTamano {
		t.Fatalf("MbrTamano = %d, want %d", got.MbrTamano, mbr.MbrTamano)
	}
	if got.MbrDskSignature != mbr.MbrDskSignature {
		t.Fatalf("MbrDskSignature = %d, want %d", got.MbrDskSignature, mbr.MbrDskSignature)
	}
	if got.DskFit != mbr.DskFit {
		t.Fatalf("DskFit = %q, want %q", got.DskFit, mbr.DskFit)
	}
	if structs.FixedBytesToString(got.MbrPartitions[0].PartName[:]) != "Part1" {
		t.Fatalf("partition name = %q, want Part1", structs.FixedBytesToString(got.MbrPartitions[0].PartName[:]))
	}

	size, err := FileSize(file)
	if err != nil {
		t.Fatalf("FileSize failed: %v", err)
	}
	if size != 1024 {
		t.Fatalf("file size = %d, want 1024", size)
	}
}

func TestWriteStructOutOfRangeFails(t *testing.T) {
	file := createFixedTempFile(t, 1024)
	defer file.Close()

	mbr := structs.NewEmptyMBR()
	size, err := BinarySize(mbr)
	if err != nil {
		t.Fatalf("BinarySize failed: %v", err)
	}

	if err := WriteStructAt(file, 1024-size+1, mbr); err == nil {
		t.Fatal("expected out-of-range error")
	}
}

func TestNegativeOffsetFails(t *testing.T) {
	file := createFixedTempFile(t, 1024)
	defer file.Close()

	if err := WriteStructAt(file, -1, structs.NewEmptyMBR()); err == nil {
		t.Fatal("expected negative offset error")
	}
}

func TestBytesReadWriteRangeValidation(t *testing.T) {
	file := createFixedTempFile(t, 1024)
	defer file.Close()

	data := []byte{1, 2, 3, 4}
	if err := WriteBytesAt(file, 100, data); err != nil {
		t.Fatalf("WriteBytesAt failed: %v", err)
	}

	got, err := ReadBytesAt(file, 100, int64(len(data)))
	if err != nil {
		t.Fatalf("ReadBytesAt failed: %v", err)
	}
	for i := range data {
		if got[i] != data[i] {
			t.Fatalf("got[%d] = %d, want %d", i, got[i], data[i])
		}
	}

	if err := WriteBytesAt(file, 1022, []byte{1, 2, 3}); err == nil {
		t.Fatal("expected out-of-range byte write error")
	}
}

func createFixedTempFile(t *testing.T, size int64) *os.File {
	t.Helper()

	path := filepath.Join(t.TempDir(), "disk.mia")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	if err := file.Truncate(size); err != nil {
		file.Close()
		t.Fatalf("Truncate failed: %v", err)
	}
	return file
}
