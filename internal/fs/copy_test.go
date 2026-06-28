package fs

import (
	"strings"
	"testing"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

func TestCopySimpleFileCreatesIndependentInodeAndBlocks(t *testing.T) {
	resetMount(t)
	diskPath := setupTransferFS(t)
	actor := rootActor()
	beforeSize := testFileSize(t, diskPath)
	if err := Mkfile(diskPath, activePartitionStart(t), "/src/a.txt", false, 150, "", actor); err != nil {
		t.Fatal(err)
	}
	sourceIndex, source := resolveEntry(t, diskPath, "/src/a.txt")

	warnings, err := Copy(diskPath, activePartitionStart(t), "/src/a.txt", "/dst", actor)
	if err != nil {
		t.Fatalf("copy file: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	copyIndex, copied := resolveEntry(t, diskPath, "/dst/a.txt")
	if copyIndex == sourceIndex {
		t.Fatal("copy reused source inode")
	}
	if copied.IBlock[0] == source.IBlock[0] {
		t.Fatal("copy reused source data block")
	}
	if got, want := readFileFromFS(t, diskPath, "/dst/a.txt"), readFileFromFS(t, diskPath, "/src/a.txt"); string(got) != string(want) {
		t.Fatal("copied content differs")
	}
	if copied.IPerm != source.IPerm {
		t.Fatalf("permissions changed: got %q want %q", copied.IPerm, source.IPerm)
	}
	if got := testFileSize(t, diskPath); got != beforeSize {
		t.Fatalf("disk size changed from %d to %d", beforeSize, got)
	}
}

func TestCopyEmptyAndIndirectFiles(t *testing.T) {
	resetMount(t)
	diskPath := setupTransferFS(t)
	actor := rootActor()
	if err := Mkfile(diskPath, activePartitionStart(t), "/src/empty.txt", false, 0, "", actor); err != nil {
		t.Fatal(err)
	}
	largeSize := int64(directBlockLimit*BlockSize + 75)
	if err := Mkfile(diskPath, activePartitionStart(t), "/src/large.txt", false, largeSize, "", actor); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/src/empty.txt", "/src/large.txt"} {
		if _, err := Copy(diskPath, activePartitionStart(t), path, "/dst", actor); err != nil {
			t.Fatalf("copy %s: %v", path, err)
		}
	}
	empty := resolveInode(t, diskPath, "/dst/empty.txt")
	if empty.ISize != 0 {
		t.Fatalf("empty copy size = %d", empty.ISize)
	}
	for _, block := range empty.IBlock {
		if block != -1 {
			t.Fatalf("empty copy has block %d", block)
		}
	}
	large := resolveInode(t, diskPath, "/dst/large.txt")
	if large.IBlock[simpleIndirectIndex] < 0 {
		t.Fatal("indirect copy has no pointer block")
	}
	if got := readFileFromFS(t, diskPath, "/dst/large.txt"); len(got) != int(largeSize) {
		t.Fatalf("large copy length = %d", len(got))
	}
}

func TestCopyRecursiveDirectoryWithMultipleBlocks(t *testing.T) {
	resetMount(t)
	diskPath := setupTransferFS(t)
	actor := rootActor()
	if err := Mkdir(diskPath, activePartitionStart(t), "/src/deep/sub", true, actor); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 10; index++ {
		name := "/src/f" + string(rune('a'+index)) + ".txt"
		if err := Mkfile(diskPath, activePartitionStart(t), name, false, int64(index*15), "", actor); err != nil {
			t.Fatal(err)
		}
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/src/deep/sub/end.txt", false, 900, "", actor); err != nil {
		t.Fatal(err)
	}

	if _, err := Copy(diskPath, activePartitionStart(t), "/src", "/dst", actor); err != nil {
		t.Fatalf("copy recursive: %v", err)
	}
	assertPathExists(t, diskPath, "/src/deep/sub/end.txt")
	assertPathExists(t, diskPath, "/dst/src/deep/sub/end.txt")
	if got := readFileFromFS(t, diskPath, "/dst/src/deep/sub/end.txt"); len(got) != 900 {
		t.Fatalf("deep copied file length = %d", len(got))
	}
	copiedDir := resolveInode(t, diskPath, "/dst/src")
	if copiedDir.IBlock[1] < 0 {
		t.Fatal("expected copied directory to use multiple folder blocks")
	}
}

func TestCopyEmptyDirectory(t *testing.T) {
	resetMount(t)
	diskPath := setupTransferFS(t)
	if err := Mkdir(diskPath, activePartitionStart(t), "/src/emptydir", false, rootActor()); err != nil {
		t.Fatal(err)
	}
	if _, err := Copy(diskPath, activePartitionStart(t), "/src/emptydir", "/dst", rootActor()); err != nil {
		t.Fatal(err)
	}
	assertPathExists(t, diskPath, "/dst/emptydir")
}

func TestCopyRejectsInvalidPathsAndDuplicates(t *testing.T) {
	resetMount(t)
	diskPath := setupTransferFS(t)
	actor := rootActor()
	if err := Mkfile(diskPath, activePartitionStart(t), "/src/a.txt", false, 10, "", actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/dst/a.txt", false, 10, "", actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/target.txt", false, 10, "", actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkdir(diskPath, activePartitionStart(t), "/src/sub", false, actor); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		source      string
		destination string
	}{
		{"/", "/dst"},
		{"/missing", "/dst"},
		{"/src/a.txt", "/missing"},
		{"/src/a.txt", "/target.txt"},
		{"/src/a.txt", "/dst"},
		{"/src", "/src/sub"},
	}
	for _, test := range cases {
		if _, err := Copy(diskPath, activePartitionStart(t), test.source, test.destination, actor); err == nil {
			t.Fatalf("expected copy error: %s -> %s", test.source, test.destination)
		}
	}
}

func TestCopySkipsUnreadableDescendant(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := Actor{User: "user2", UID: 2, GID: 2}
	if err := Mkdir(diskPath, activePartitionStart(t), "/src", false, actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkdir(diskPath, activePartitionStart(t), "/dst", false, actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/src/ok.txt", false, 10, "", actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/src/locked.txt", false, 10, "", actor); err != nil {
		t.Fatal(err)
	}
	setEntryOwnerAndPerm(t, diskPath, "/src/locked.txt", 3, 3, "000")

	warnings, err := Copy(diskPath, activePartitionStart(t), "/src", "/dst", actor)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "locked.txt") {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	assertPathExists(t, diskPath, "/dst/src/ok.txt")
	assertPathMissing(t, diskPath, "/dst/src/locked.txt")
}

func setupTransferFS(t *testing.T) string {
	t.Helper()
	diskPath := setupFormattedFS(t)
	actor := rootActor()
	for _, path := range []string{"/src", "/dst", "/other"} {
		if err := Mkdir(diskPath, activePartitionStart(t), path, false, actor); err != nil {
			t.Fatal(err)
		}
	}
	return diskPath
}

func resolveEntry(t *testing.T, diskPath string, path string) (int32, structs.Inode) {
	t.Helper()
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sb := activeSuperBlock(t, file)
	index, inode, err := ResolvePath(file, sb, path)
	if err != nil {
		t.Fatal(err)
	}
	return index, inode
}

func setEntryOwnerAndPerm(t *testing.T, diskPath string, path string, uid int32, gid int32, perm string) {
	t.Helper()
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sb := activeSuperBlock(t, file)
	index, inode, err := ResolvePath(file, sb, path)
	if err != nil {
		t.Fatal(err)
	}
	inode.IUid = uid
	inode.IGid = gid
	structs.SetPerm(&inode.IPerm, perm)
	if err := WriteInode(file, sb, index, inode); err != nil {
		t.Fatal(err)
	}
}
