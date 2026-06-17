package fs

import (
	"strings"
	"testing"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/mount"
)

func TestWriteRootUsersFileGrowsAndShrinksBlocks(t *testing.T) {
	resetMount(t)
	path := createMountedPartition(t, 10, 5, "P")
	var noop strings.Builder
	if err := FormatFromParams(map[string]string{"id": "961A"}, &noop); err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	beforeSize := testFileSize(t, path)

	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()

	mounted, _ := mount.Global.GetMounted("961A")
	sb, err := ReadSuperBlock(file, int64(mounted.Start))
	if err != nil {
		t.Fatalf("ReadSuperBlock failed: %v", err)
	}
	beforeFree := sb.SFreeBlocksCount
	largeContent := "1,G,root\n1,U,root,root,123\n" + strings.Repeat("2,G,grupo\n", 8)
	if len(largeContent) <= BlockSize {
		t.Fatalf("test content should exceed one block")
	}

	sb, err = WriteRootUsersFile(file, sb, largeContent)
	if err != nil {
		t.Fatalf("WriteRootUsersFile large failed: %v", err)
	}
	got, err := ReadRootUsersFile(file, sb)
	if err != nil {
		t.Fatalf("ReadRootUsersFile failed: %v", err)
	}
	if got != largeContent {
		t.Fatalf("content mismatch after grow")
	}
	if sb.SFreeBlocksCount >= beforeFree {
		t.Fatalf("expected fewer free blocks after grow, before=%d after=%d", beforeFree, sb.SFreeBlocksCount)
	}
	afterGrowFree := sb.SFreeBlocksCount
	blockBitmap, err := ReadBitmap(file, int64(sb.SBmBlockStart), sb.SBlocksCount)
	if err != nil {
		t.Fatalf("ReadBitmap failed: %v", err)
	}
	if blockBitmap[2] != 1 {
		t.Fatalf("expected block 2 allocated, bitmap prefix=%v", blockBitmap[:4])
	}

	smallContent := "1,G,root\n1,U,root,root,123\n"
	sb, err = WriteRootUsersFile(file, sb, smallContent)
	if err != nil {
		t.Fatalf("WriteRootUsersFile small failed: %v", err)
	}
	got, err = ReadRootUsersFile(file, sb)
	if err != nil {
		t.Fatalf("ReadRootUsersFile failed: %v", err)
	}
	if got != smallContent {
		t.Fatalf("content mismatch after shrink")
	}
	if sb.SFreeBlocksCount <= afterGrowFree {
		t.Fatalf("expected more free blocks after shrink, grow=%d after=%d", afterGrowFree, sb.SFreeBlocksCount)
	}
	blockBitmap, err = ReadBitmap(file, int64(sb.SBmBlockStart), sb.SBlocksCount)
	if err != nil {
		t.Fatalf("ReadBitmap failed: %v", err)
	}
	if blockBitmap[2] != 0 {
		t.Fatalf("expected block 2 freed, bitmap prefix=%v", blockBitmap[:4])
	}

	if afterSize := testFileSize(t, path); afterSize != beforeSize {
		t.Fatalf("disk size changed from %d to %d", beforeSize, afterSize)
	}
}

func TestWriteRootUsersFileRejectsTooManyDirectBlocks(t *testing.T) {
	resetMount(t)
	path := createMountedPartition(t, 10, 5, "P")
	var noop strings.Builder
	if err := FormatFromParams(map[string]string{"id": "961A"}, &noop); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()
	mounted, _ := mount.Global.GetMounted("961A")
	sb, err := ReadSuperBlock(file, int64(mounted.Start))
	if err != nil {
		t.Fatalf("ReadSuperBlock failed: %v", err)
	}
	content := strings.Repeat("x", directBlockLimit*BlockSize+1)
	if _, err := WriteRootUsersFile(file, sb, content); err == nil {
		t.Fatal("expected direct block capacity error")
	}
}

func TestAllocateBlocksRollback(t *testing.T) {
	bitmap := []byte{1, 0}
	if _, err := AllocateBlocks(bitmap, 2); err == nil {
		t.Fatal("expected allocation error")
	}
	if bitmap[1] != 0 {
		t.Fatalf("expected rollback, bitmap=%v", bitmap)
	}
}

func TestWriteRootUsersFileDoesNotChangeDiskSize(t *testing.T) {
	resetMount(t)
	path := createMountedPartition(t, 10, 5, "P")
	var noop strings.Builder
	if err := FormatFromParams(map[string]string{"id": "961A"}, &noop); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()
	before, err := binio.FileSize(file)
	if err != nil {
		t.Fatalf("FileSize failed: %v", err)
	}
	mounted, _ := mount.Global.GetMounted("961A")
	sb, err := ReadSuperBlock(file, int64(mounted.Start))
	if err != nil {
		t.Fatalf("ReadSuperBlock failed: %v", err)
	}
	if _, err := WriteRootUsersFile(file, sb, "1,G,root\n1,U,root,root,123\n2,G,users\n"); err != nil {
		t.Fatalf("WriteRootUsersFile failed: %v", err)
	}
	after, err := binio.FileSize(file)
	if err != nil {
		t.Fatalf("FileSize failed: %v", err)
	}
	if after != before {
		t.Fatalf("file grew from %d to %d", before, after)
	}
}
