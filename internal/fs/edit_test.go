package fs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

func TestEditFileSmallerFreesIndirectPointer(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := rootActor()
	beforeSize := testFileSize(t, diskPath)
	largeSize := int64(directBlockLimit*BlockSize + 75)
	if err := Mkfile(diskPath, activePartitionStart(t), "/large.txt", false, largeSize, "", actor); err != nil {
		t.Fatal(err)
	}
	beforeSB := readSBByPath(t, diskPath)
	beforeInode := resolveInode(t, diskPath, "/large.txt")
	pointerIndex := beforeInode.IBlock[simpleIndirectIndex]
	if pointerIndex < 0 {
		t.Fatal("expected indirect pointer before edit")
	}

	contentPath := writeHostContent(t, []byte("short"))
	if err := Edit(diskPath, activePartitionStart(t), "/large.txt", contentPath, actor); err != nil {
		t.Fatalf("edit smaller: %v", err)
	}
	if got := string(readFileFromFS(t, diskPath, "/large.txt")); got != "short" {
		t.Fatalf("edited content = %q", got)
	}
	afterInode := resolveInode(t, diskPath, "/large.txt")
	if afterInode.IBlock[simpleIndirectIndex] != -1 {
		t.Fatalf("pointer block was not released: %d", afterInode.IBlock[simpleIndirectIndex])
	}
	afterSB := readSBByPath(t, diskPath)
	if afterSB.SFreeBlocksCount <= beforeSB.SFreeBlocksCount {
		t.Fatal("expected blocks to be released")
	}
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	bitmap, err := ReadBlockBitmap(file, afterSB)
	if err != nil {
		t.Fatal(err)
	}
	if bitmap[pointerIndex] != 0 {
		t.Fatal("pointer block remains marked as used")
	}
	if got := testFileSize(t, diskPath); got != beforeSize {
		t.Fatalf("disk size changed from %d to %d", beforeSize, got)
	}
}

func TestEditFileLargerUsesSimpleIndirect(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := rootActor()
	if err := Mkfile(diskPath, activePartitionStart(t), "/grow.txt", false, 10, "", actor); err != nil {
		t.Fatal(err)
	}
	content := []byte(strings.Repeat("MIA", 300))
	contentPath := writeHostContent(t, content)

	if err := Edit(diskPath, activePartitionStart(t), "/grow.txt", contentPath, actor); err != nil {
		t.Fatalf("edit larger: %v", err)
	}
	if got := readFileFromFS(t, diskPath, "/grow.txt"); string(got) != string(content) {
		t.Fatalf("large edited content differs, got %d bytes", len(got))
	}
	inode := resolveInode(t, diskPath, "/grow.txt")
	if inode.IBlock[simpleIndirectIndex] < 0 {
		t.Fatal("expected simple indirect pointer after edit")
	}
}

func TestEditFileToEmptyReleasesAllBlocks(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := rootActor()
	if err := Mkfile(diskPath, activePartitionStart(t), "/empty.txt", false, 150, "", actor); err != nil {
		t.Fatal(err)
	}
	contentPath := writeHostContent(t, nil)
	if err := Edit(diskPath, activePartitionStart(t), "/empty.txt", contentPath, actor); err != nil {
		t.Fatalf("edit empty: %v", err)
	}
	inode := resolveInode(t, diskPath, "/empty.txt")
	if inode.ISize != 0 {
		t.Fatalf("empty file size = %d", inode.ISize)
	}
	for index, block := range inode.IBlock {
		if block != -1 {
			t.Fatalf("IBlock[%d] = %d, want -1", index, block)
		}
	}
}

func TestEditRejectsInvalidTargetsAndContent(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := rootActor()
	if err := Mkdir(diskPath, activePartitionStart(t), "/docs", false, actor); err != nil {
		t.Fatal(err)
	}
	contentPath := writeHostContent(t, []byte("new"))

	if err := Edit(diskPath, activePartitionStart(t), "/docs", contentPath, actor); err == nil {
		t.Fatal("expected directory edit error")
	}
	if err := Edit(diskPath, activePartitionStart(t), "/missing.txt", contentPath, actor); err == nil {
		t.Fatal("expected missing EXT2 file error")
	}
	if err := Edit(diskPath, activePartitionStart(t), "/users.txt", filepath.Join(t.TempDir(), "missing.txt"), actor); err == nil {
		t.Fatal("expected missing host file error")
	}
}

func TestEditRequiresReadAndWritePermission(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	root := rootActor()
	if err := Mkfile(diskPath, activePartitionStart(t), "/locked.txt", false, 10, "", root); err != nil {
		t.Fatal(err)
	}
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	sb := activeSuperBlock(t, file)
	index, inode, err := ResolvePath(file, sb, "/locked.txt")
	if err != nil {
		file.Close()
		t.Fatal(err)
	}
	structs.SetPerm(&inode.IPerm, "400")
	if err := WriteInode(file, sb, index, inode); err != nil {
		file.Close()
		t.Fatal(err)
	}
	file.Close()

	contentPath := writeHostContent(t, []byte("denied"))
	if err := Edit(diskPath, activePartitionStart(t), "/locked.txt", contentPath, Actor{User: "user", UID: 2, GID: 2}); err == nil {
		t.Fatal("expected edit permission error")
	}
	if err := Edit(diskPath, activePartitionStart(t), "/locked.txt", contentPath, root); err != nil {
		t.Fatalf("root edit failed: %v", err)
	}
}

func writeHostContent(t *testing.T, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "content.txt")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
