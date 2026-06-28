package fs

import (
	"testing"

	"mia_p1_201800996/internal/disk"
)

func TestMoveFileKeepsInodeBlocksAndFreeCounts(t *testing.T) {
	resetMount(t)
	diskPath := setupTransferFS(t)
	actor := rootActor()
	beforeSize := testFileSize(t, diskPath)
	if err := Mkfile(diskPath, activePartitionStart(t), "/src/a.txt", false, 150, "", actor); err != nil {
		t.Fatal(err)
	}
	sourceIndex, source := resolveEntry(t, diskPath, "/src/a.txt")
	beforeSB := readSBByPath(t, diskPath)

	if err := Move(diskPath, activePartitionStart(t), "/src/a.txt", "/dst", actor); err != nil {
		t.Fatalf("move file: %v", err)
	}
	assertPathMissing(t, diskPath, "/src/a.txt")
	movedIndex, moved := resolveEntry(t, diskPath, "/dst/a.txt")
	if movedIndex != sourceIndex {
		t.Fatalf("inode changed from %d to %d", sourceIndex, movedIndex)
	}
	if moved.IBlock != source.IBlock || moved.ISize != source.ISize {
		t.Fatal("move changed file blocks or size")
	}
	afterSB := readSBByPath(t, diskPath)
	if afterSB.SFreeInodesCount != beforeSB.SFreeInodesCount || afterSB.SFreeBlocksCount != beforeSB.SFreeBlocksCount {
		t.Fatalf("move allocated or freed resources: before=%#v after=%#v", beforeSB, afterSB)
	}
	if got := testFileSize(t, diskPath); got != beforeSize {
		t.Fatalf("disk size changed from %d to %d", beforeSize, got)
	}
}

func TestMoveDeepDirectoryUpdatesParentReference(t *testing.T) {
	resetMount(t)
	diskPath := setupTransferFS(t)
	actor := rootActor()
	if err := Mkdir(diskPath, activePartitionStart(t), "/src/folder/deep", true, actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/src/folder/deep/a.txt", false, 20, "", actor); err != nil {
		t.Fatal(err)
	}
	sourceIndex, _ := resolveEntry(t, diskPath, "/src/folder")
	destinationIndex, _ := resolveEntry(t, diskPath, "/dst")

	if err := Move(diskPath, activePartitionStart(t), "/src/folder", "/dst", actor); err != nil {
		t.Fatalf("move directory: %v", err)
	}
	assertPathMissing(t, diskPath, "/src/folder")
	movedIndex, moved := resolveEntry(t, diskPath, "/dst/folder")
	if movedIndex != sourceIndex {
		t.Fatalf("directory inode changed from %d to %d", sourceIndex, movedIndex)
	}
	assertPathExists(t, diskPath, "/dst/folder/deep/a.txt")

	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sb := activeSuperBlock(t, file)
	block, err := ReadFolderBlock(file, sb, moved.IBlock[0])
	if err != nil {
		t.Fatal(err)
	}
	if block.BContent[1].BInodo != destinationIndex {
		t.Fatalf(".. points to %d, want %d", block.BContent[1].BInodo, destinationIndex)
	}
}

func TestMoveRejectsInvalidTargetsAndDuplicates(t *testing.T) {
	resetMount(t)
	diskPath := setupTransferFS(t)
	actor := rootActor()
	if err := Mkdir(diskPath, activePartitionStart(t), "/src/sub", false, actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/src/a.txt", false, 10, "", actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/dst/a.txt", false, 10, "", actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/target.txt", false, 10, "", actor); err != nil {
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
		if err := Move(diskPath, activePartitionStart(t), test.source, test.destination, actor); err == nil {
			t.Fatalf("expected move error: %s -> %s", test.source, test.destination)
		}
	}
}

func TestMoveRejectsPermissions(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := Actor{User: "user2", UID: 2, GID: 2}
	if err := Mkdir(diskPath, activePartitionStart(t), "/src", false, actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkdir(diskPath, activePartitionStart(t), "/dst", false, actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/src/a.txt", false, 10, "", actor); err != nil {
		t.Fatal(err)
	}
	setEntryOwnerAndPerm(t, diskPath, "/src/a.txt", 3, 3, "444")
	if err := Move(diskPath, activePartitionStart(t), "/src/a.txt", "/dst", actor); err == nil {
		t.Fatal("expected source permission error")
	}
	setEntryOwnerAndPerm(t, diskPath, "/src/a.txt", 2, 2, "664")
	setEntryOwnerAndPerm(t, diskPath, "/dst", 3, 3, "444")
	if err := Move(diskPath, activePartitionStart(t), "/src/a.txt", "/dst", actor); err == nil {
		t.Fatal("expected destination permission error")
	}
}

func TestMoveRejectsDestinationWithoutExistingFreeSlot(t *testing.T) {
	resetMount(t)
	diskPath := setupTransferFS(t)
	actor := rootActor()
	if err := Mkfile(diskPath, activePartitionStart(t), "/src/move.txt", false, 10, "", actor); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.txt", "b.txt"} {
		if err := Mkfile(diskPath, activePartitionStart(t), "/dst/"+name, false, 0, "", actor); err != nil {
			t.Fatal(err)
		}
	}
	if err := Move(diskPath, activePartitionStart(t), "/src/move.txt", "/dst", actor); err == nil {
		t.Fatal("expected move to reject block allocation")
	}
}
