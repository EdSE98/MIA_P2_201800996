package fs

import (
	"strings"
	"testing"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

func TestRemoveSimpleAndEmptyFiles(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := rootActor()
	beforeSize := testFileSize(t, diskPath)
	if err := Mkfile(diskPath, activePartitionStart(t), "/simple.txt", false, 20, "", actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/empty.txt", false, 0, "", actor); err != nil {
		t.Fatal(err)
	}

	if err := Remove(diskPath, activePartitionStart(t), "/simple.txt", actor); err != nil {
		t.Fatalf("remove simple: %v", err)
	}
	if err := Remove(diskPath, activePartitionStart(t), "/empty.txt", actor); err != nil {
		t.Fatalf("remove empty: %v", err)
	}
	assertPathMissing(t, diskPath, "/simple.txt")
	assertPathMissing(t, diskPath, "/empty.txt")
	if got := testFileSize(t, diskPath); got != beforeSize {
		t.Fatalf("disk size changed from %d to %d", beforeSize, got)
	}
}

func TestRemoveDirectFileUpdatesBitmaps(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := rootActor()
	if err := Mkfile(diskPath, activePartitionStart(t), "/direct.txt", false, 150, "", actor); err != nil {
		t.Fatal(err)
	}
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	sbBefore := activeSuperBlock(t, file)
	inodeIndex, inode, err := ResolvePath(file, sbBefore, "/direct.txt")
	if err != nil {
		file.Close()
		t.Fatal(err)
	}
	blocks, err := FileDataBlockIndices(file, sbBefore, inode)
	file.Close()
	if err != nil {
		t.Fatal(err)
	}

	if err := Remove(diskPath, activePartitionStart(t), "/direct.txt", actor); err != nil {
		t.Fatal(err)
	}
	file, _, err = disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sbAfter := activeSuperBlock(t, file)
	if sbAfter.SFreeBlocksCount != sbBefore.SFreeBlocksCount+int32(len(blocks)) {
		t.Fatalf("free blocks = %d, want %d", sbAfter.SFreeBlocksCount, sbBefore.SFreeBlocksCount+int32(len(blocks)))
	}
	if sbAfter.SFreeInodesCount != sbBefore.SFreeInodesCount+1 {
		t.Fatalf("free inodes = %d, want %d", sbAfter.SFreeInodesCount, sbBefore.SFreeInodesCount+1)
	}
	inodeBitmap, err := ReadInodeBitmap(file, sbAfter)
	if err != nil {
		t.Fatal(err)
	}
	if inodeBitmap[inodeIndex] != 0 {
		t.Fatal("removed inode remains occupied")
	}
	if sbAfter.SFirstIno != RecalculateFirstFree(inodeBitmap) {
		t.Fatalf("SFirstIno = %d, want %d", sbAfter.SFirstIno, RecalculateFirstFree(inodeBitmap))
	}
	blockBitmap, err := ReadBlockBitmap(file, sbAfter)
	if err != nil {
		t.Fatal(err)
	}
	for _, blockIndex := range blocks {
		if blockBitmap[blockIndex] != 0 {
			t.Fatalf("block %d remains occupied", blockIndex)
		}
	}
	if sbAfter.SFirstBlo != RecalculateFirstFree(blockBitmap) {
		t.Fatalf("SFirstBlo = %d, want %d", sbAfter.SFirstBlo, RecalculateFirstFree(blockBitmap))
	}
}

func TestRemoveFileWithSimpleIndirectPointer(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := rootActor()
	size := int64(directBlockLimit*BlockSize + 75)
	if err := Mkfile(diskPath, activePartitionStart(t), "/large.txt", false, size, "", actor); err != nil {
		t.Fatal(err)
	}
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	sbBefore := activeSuperBlock(t, file)
	_, inode, err := ResolvePath(file, sbBefore, "/large.txt")
	if err != nil {
		file.Close()
		t.Fatal(err)
	}
	dataBlocks, err := FileDataBlockIndices(file, sbBefore, inode)
	file.Close()
	if err != nil {
		t.Fatal(err)
	}
	pointerIndex := inode.IBlock[simpleIndirectIndex]
	if pointerIndex < 0 {
		t.Fatal("expected indirect pointer")
	}

	if err := Remove(diskPath, activePartitionStart(t), "/large.txt", actor); err != nil {
		t.Fatal(err)
	}
	file, _, err = disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sbAfter := activeSuperBlock(t, file)
	blockBitmap, err := ReadBlockBitmap(file, sbAfter)
	if err != nil {
		t.Fatal(err)
	}
	for _, blockIndex := range append(dataBlocks, pointerIndex) {
		if blockBitmap[blockIndex] != 0 {
			t.Fatalf("block %d remains occupied", blockIndex)
		}
	}
	wantFreed := int32(len(dataBlocks) + 1)
	if sbAfter.SFreeBlocksCount != sbBefore.SFreeBlocksCount+wantFreed {
		t.Fatalf("free blocks = %d, want %d", sbAfter.SFreeBlocksCount, sbBefore.SFreeBlocksCount+wantFreed)
	}
}

func TestRemoveEmptyDirectory(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := rootActor()
	if err := Mkdir(diskPath, activePartitionStart(t), "/emptydir", false, actor); err != nil {
		t.Fatal(err)
	}
	if err := Remove(diskPath, activePartitionStart(t), "/emptydir", actor); err != nil {
		t.Fatal(err)
	}
	assertPathMissing(t, diskPath, "/emptydir")
}

func TestRemoveReleasesEmptyAdditionalFolderBlock(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := rootActor()
	if err := Mkdir(diskPath, activePartitionStart(t), "/many", false, actor); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if err := Mkfile(diskPath, activePartitionStart(t), "/many/"+name, false, 0, "", actor); err != nil {
			t.Fatal(err)
		}
	}
	before := resolveInode(t, diskPath, "/many")
	extraBlock := before.IBlock[1]
	if extraBlock < 0 {
		t.Fatal("expected additional folder block")
	}

	if err := Remove(diskPath, activePartitionStart(t), "/many/c.txt", actor); err != nil {
		t.Fatal(err)
	}
	after := resolveInode(t, diskPath, "/many")
	if after.IBlock[1] != -1 {
		t.Fatalf("empty folder block still referenced: %d", after.IBlock[1])
	}
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sb := activeSuperBlock(t, file)
	bitmap, err := ReadBlockBitmap(file, sb)
	if err != nil {
		t.Fatal(err)
	}
	if bitmap[extraBlock] != 0 {
		t.Fatal("empty folder block remains occupied")
	}
}

func TestRemoveDeepDirectoryWithMultipleFolderBlocks(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := rootActor()
	baseline := readSBByPath(t, diskPath)
	if err := Mkdir(diskPath, activePartitionStart(t), "/tree/deep/sub", true, actor); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 10; index++ {
		name := "/tree/f" + string(rune('a'+index)) + ".txt"
		if err := Mkfile(diskPath, activePartitionStart(t), name, false, int64(index*20), "", actor); err != nil {
			t.Fatal(err)
		}
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/tree/deep/sub/end.txt", false, 900, "", actor); err != nil {
		t.Fatal(err)
	}

	if err := Remove(diskPath, activePartitionStart(t), "/tree", actor); err != nil {
		t.Fatalf("remove recursive tree: %v", err)
	}
	assertPathMissing(t, diskPath, "/tree")
	after := readSBByPath(t, diskPath)
	if after.SFreeBlocksCount != baseline.SFreeBlocksCount || after.SFreeInodesCount != baseline.SFreeInodesCount {
		t.Fatalf("resources not fully restored: before=%#v after=%#v", baseline, after)
	}

	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sb := activeSuperBlock(t, file)
	_, root, err := ResolvePath(file, sb, "/")
	if err != nil {
		t.Fatal(err)
	}
	items, err := ListDirectoryEntries(file, sb, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		if item.Name == "tree" {
			t.Fatal("visualizer still lists removed directory")
		}
	}
}

func TestRemoveRejectsRootAndMissingPath(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	if err := Remove(diskPath, activePartitionStart(t), "/", rootActor()); err == nil {
		t.Fatal("expected root removal error")
	}
	if err := Remove(diskPath, activePartitionStart(t), "/missing", rootActor()); err == nil {
		t.Fatal("expected missing path error")
	}
}

func TestRemovePermissionBlockKeepsDirectoryAndLaterChild(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := Actor{User: "user2", UID: 2, GID: 2}
	if err := Mkdir(diskPath, activePartitionStart(t), "/owned", false, actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/owned/ok.txt", false, 20, "", actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/owned/locked.txt", false, 20, "", actor); err != nil {
		t.Fatal(err)
	}

	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	sb := activeSuperBlock(t, file)
	lockedIndex, locked, err := ResolvePath(file, sb, "/owned/locked.txt")
	if err != nil {
		file.Close()
		t.Fatal(err)
	}
	locked.IUid = 3
	locked.IGid = 3
	structs.SetPerm(&locked.IPerm, "444")
	if err := WriteInode(file, sb, lockedIndex, locked); err != nil {
		file.Close()
		t.Fatal(err)
	}
	file.Close()

	err = Remove(diskPath, activePartitionStart(t), "/owned", actor)
	if err == nil || !strings.Contains(err.Error(), "permiso") {
		t.Fatalf("expected permission error, got %v", err)
	}
	assertPathMissing(t, diskPath, "/owned/ok.txt")
	assertPathExists(t, diskPath, "/owned")
	assertPathExists(t, diskPath, "/owned/locked.txt")
}

func assertPathMissing(t *testing.T, diskPath string, path string) {
	t.Helper()
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sb := activeSuperBlock(t, file)
	if _, _, err := ResolvePath(file, sb, path); err == nil {
		t.Fatalf("expected path %q to be missing", path)
	}
}

func assertPathExists(t *testing.T, diskPath string, path string) {
	t.Helper()
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sb := activeSuperBlock(t, file)
	if _, _, err := ResolvePath(file, sb, path); err != nil {
		t.Fatalf("expected path %q to exist: %v", path, err)
	}
}
