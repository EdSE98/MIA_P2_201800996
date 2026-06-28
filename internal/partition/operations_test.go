package partition

import (
	"bytes"
	"path/filepath"
	"testing"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

func TestResizePrimaryGrowAndShrink(t *testing.T) {
	path := makeOperationDisk(t, "resize-primary.dsk", 4)
	createOperationPartition(t, path, "Part1", 1, "M", "P")

	if err := Resize(ResizeOptions{Path: path, Name: "Part1", Add: 512, Unit: "K"}); err != nil {
		t.Fatalf("grow partition: %v", err)
	}
	part, _, err := SearchPartition(path, "Part1")
	if err != nil {
		t.Fatal(err)
	}
	if part.PartSize != 1536*1024 {
		t.Fatalf("size after grow = %d, want %d", part.PartSize, 1536*1024)
	}

	if err := Resize(ResizeOptions{Path: path, Name: "Part1", Add: -256, Unit: "K"}); err != nil {
		t.Fatalf("shrink partition: %v", err)
	}
	part, _, err = SearchPartition(path, "Part1")
	if err != nil {
		t.Fatal(err)
	}
	if part.PartSize != 1280*1024 {
		t.Fatalf("size after shrink = %d, want %d", part.PartSize, 1280*1024)
	}
	assertDiskSize(t, path, 4*1024*1024)
}

func TestResizeRejectsNonPositiveAndMissingContiguousSpace(t *testing.T) {
	path := makeOperationDisk(t, "resize-limits.dsk", 4)
	createOperationPartition(t, path, "Part1", 1, "M", "P")
	createOperationPartition(t, path, "Part2", 1, "M", "P")

	if err := Resize(ResizeOptions{Path: path, Name: "Part1", Add: 1, Unit: "K"}); err == nil {
		t.Fatal("expected contiguous space error")
	}
	if err := Resize(ResizeOptions{Path: path, Name: "Part2", Add: -1, Unit: "M"}); err == nil {
		t.Fatal("expected non-positive size error")
	}
	assertDiskSize(t, path, 4*1024*1024)
}

func TestResizeRejectsFormattedPartitionShrink(t *testing.T) {
	path := makeOperationDisk(t, "resize-formatted.dsk", 4)
	createOperationPartition(t, path, "Part1", 1, "M", "P")
	part, _, err := SearchPartition(path, "Part1")
	if err != nil {
		t.Fatal(err)
	}

	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatal(err)
	}
	sb := structs.SuperBlock{SFilesystemType: 2, SMagic: 0xEF53}
	if err := binio.WriteStructAt(file, int64(part.PartStart), sb); err != nil {
		file.Close()
		t.Fatal(err)
	}
	file.Close()

	if err := Resize(ResizeOptions{Path: path, Name: "Part1", Add: -1, Unit: "K"}); err == nil {
		t.Fatal("expected formatted partition shrink error")
	}
}

func TestResizeLogicalWithinExtendedRange(t *testing.T) {
	path := makeOperationDisk(t, "resize-logical.dsk", 4)
	createOperationPartition(t, path, "Ext1", 2, "M", "E")
	createOperationPartition(t, path, "Log1", 256, "K", "L")

	if err := Resize(ResizeOptions{Path: path, Name: "Log1", Add: 128, Unit: "K"}); err != nil {
		t.Fatalf("grow logical: %v", err)
	}
	if err := Resize(ResizeOptions{Path: path, Name: "Log1", Add: -64, Unit: "K"}); err != nil {
		t.Fatalf("shrink logical: %v", err)
	}

	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	mbr, err := disk.ReadMBR(path)
	if err != nil {
		t.Fatal(err)
	}
	extended, _, ok := FindExtendedPartition(mbr)
	if !ok {
		t.Fatal("expected extended partition")
	}
	chain, err := ReadEBRChain(file, extended)
	if err != nil {
		t.Fatal(err)
	}
	if chain[0].EBR.PartSize != 320*1024 {
		t.Fatalf("logical size = %d, want %d", chain[0].EBR.PartSize, 320*1024)
	}
}

func TestResizeLogicalRejectsGrowthIntoNextEBR(t *testing.T) {
	path := makeOperationDisk(t, "resize-logical-boundary.dsk", 4)
	createOperationPartition(t, path, "Ext1", 2, "M", "E")
	createOperationPartition(t, path, "Log1", 256, "K", "L")
	createOperationPartition(t, path, "Log2", 256, "K", "L")

	if err := Resize(ResizeOptions{Path: path, Name: "Log1", Add: 1, Unit: "B"}); err == nil {
		t.Fatal("expected logical contiguous space error")
	}
}

func TestDeletePrimaryFastPreservesData(t *testing.T) {
	path := makeOperationDisk(t, "delete-fast.dsk", 2)
	createOperationPartition(t, path, "Part1", 256, "K", "P")
	part, _, err := SearchPartition(path, "Part1")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("MIA-DATA")
	writeAt(t, path, int64(part.PartStart), marker)

	if err := Delete(DeleteOptions{Path: path, Name: "Part1", Mode: "fast"}); err != nil {
		t.Fatalf("fast delete: %v", err)
	}
	if _, _, err := SearchPartition(path, "Part1"); err == nil {
		t.Fatal("expected partition to be removed from MBR")
	}
	if got := readAt(t, path, int64(part.PartStart), int64(len(marker))); !bytes.Equal(got, marker) {
		t.Fatalf("fast delete changed partition data: %q", got)
	}
}

func TestDeletePrimaryFullClearsData(t *testing.T) {
	path := makeOperationDisk(t, "delete-full.dsk", 2)
	createOperationPartition(t, path, "Part1", 128, "K", "P")
	part, _, err := SearchPartition(path, "Part1")
	if err != nil {
		t.Fatal(err)
	}
	writeAt(t, path, int64(part.PartStart), bytes.Repeat([]byte{0x7f}, int(part.PartSize)))

	if err := Delete(DeleteOptions{Path: path, Name: "Part1", Mode: "full"}); err != nil {
		t.Fatalf("full delete: %v", err)
	}
	cleared := readAt(t, path, int64(part.PartStart), int64(part.PartSize))
	if !bytes.Equal(cleared, make([]byte, len(cleared))) {
		t.Fatal("full delete did not clear the complete partition range")
	}
	assertDiskSize(t, path, 2*1024*1024)
}

func TestDeleteExtendedFullRemovesLogicalPartitions(t *testing.T) {
	path := makeOperationDisk(t, "delete-extended.dsk", 4)
	createOperationPartition(t, path, "Ext1", 2, "M", "E")
	createOperationPartition(t, path, "Log1", 256, "K", "L")
	createOperationPartition(t, path, "Log2", 256, "K", "L")
	extended, _, err := SearchPartition(path, "Ext1")
	if err != nil {
		t.Fatal(err)
	}

	if err := Delete(DeleteOptions{Path: path, Name: "Ext1", Mode: "full"}); err != nil {
		t.Fatalf("delete extended: %v", err)
	}
	if _, _, err := SearchPartition(path, "Ext1"); err == nil {
		t.Fatal("expected extended partition to be removed")
	}
	cleared := readAt(t, path, int64(extended.PartStart), int64(extended.PartSize))
	if !bytes.Equal(cleared, make([]byte, len(cleared))) {
		t.Fatal("extended range still contains logical partition data")
	}
}

func TestDeleteFirstLogicalRelinksEBRChain(t *testing.T) {
	path := makeOperationDisk(t, "delete-logical.dsk", 4)
	createOperationPartition(t, path, "Ext1", 2, "M", "E")
	createOperationPartition(t, path, "Log1", 256, "K", "L")
	createOperationPartition(t, path, "Log2", 256, "K", "L")
	extended, _, err := SearchPartition(path, "Ext1")
	if err != nil {
		t.Fatal(err)
	}

	if err := Delete(DeleteOptions{Path: path, Name: "Log1", Mode: "fast"}); err != nil {
		t.Fatalf("delete first logical: %v", err)
	}
	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	chain, err := ReadEBRChain(file, extended)
	if err != nil {
		t.Fatal(err)
	}
	if len(chain) != 1 {
		t.Fatalf("chain length = %d, want 1", len(chain))
	}
	if got := structs.FixedBytesToString(chain[0].EBR.PartName[:]); got != "Log2" {
		t.Fatalf("remaining logical = %q, want Log2", got)
	}
}

func TestDeleteRejectsMissingPartition(t *testing.T) {
	path := makeOperationDisk(t, "delete-missing.dsk", 1)
	if err := Delete(DeleteOptions{Path: path, Name: "Missing", Mode: "fast"}); err == nil {
		t.Fatal("expected missing partition error")
	}
}

func TestDeleteRejectsInvalidMode(t *testing.T) {
	path := makeOperationDisk(t, "delete-mode.dsk", 1)
	createOperationPartition(t, path, "Part1", 128, "K", "P")
	if err := Delete(DeleteOptions{Path: path, Name: "Part1", Mode: "quick"}); err == nil {
		t.Fatal("expected invalid delete mode error")
	}
}

func makeOperationDisk(t *testing.T, name string, sizeMB int64) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := disk.MakeDisk(disk.MakeDiskOptions{Size: sizeMB, Unit: "M", Path: path}); err != nil {
		t.Fatalf("make disk: %v", err)
	}
	return path
}

func createOperationPartition(t *testing.T, path, name string, size int64, unit, partType string) {
	t.Helper()
	if err := Create(CreateOptions{
		Size: size,
		Unit: unit,
		Path: path,
		Type: partType,
		Name: name,
	}); err != nil {
		t.Fatalf("create %s: %v", name, err)
	}
}

func writeAt(t *testing.T, path string, offset int64, data []byte) {
	t.Helper()
	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := binio.WriteBytesAt(file, offset, data); err != nil {
		t.Fatal(err)
	}
}

func readAt(t *testing.T, path string, offset, size int64) []byte {
	t.Helper()
	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	data, err := binio.ReadBytesAt(file, offset, size)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
