package fs

import (
	"strings"
	"testing"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

func TestRenameFileKeepsInodeAndContent(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := rootActor()
	beforeSize := testFileSize(t, diskPath)
	if err := Mkfile(diskPath, activePartitionStart(t), "/a.txt", false, 20, "", actor); err != nil {
		t.Fatal(err)
	}
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	sb := activeSuperBlock(t, file)
	oldIndex, oldInode, err := ResolvePath(file, sb, "/a.txt")
	file.Close()
	if err != nil {
		t.Fatal(err)
	}

	if err := Rename(diskPath, activePartitionStart(t), "/a.txt", "b1.txt", actor); err != nil {
		t.Fatalf("rename file: %v", err)
	}
	file, _, err = disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sb = activeSuperBlock(t, file)
	if _, _, err := ResolvePath(file, sb, "/a.txt"); err == nil {
		t.Fatal("old path still resolves")
	}
	newIndex, newInode, err := ResolvePath(file, sb, "/b1.txt")
	if err != nil {
		t.Fatal(err)
	}
	if newIndex != oldIndex {
		t.Fatalf("inode changed from %d to %d", oldIndex, newIndex)
	}
	if newInode.ISize != oldInode.ISize || newInode.IBlock != oldInode.IBlock {
		t.Fatal("rename changed file size or block pointers")
	}
	if got := testFileSize(t, diskPath); got != beforeSize {
		t.Fatalf("disk size changed from %d to %d", beforeSize, got)
	}
}

func TestRenameDirectoryKeepsChildren(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := rootActor()
	if err := Mkdir(diskPath, activePartitionStart(t), "/docs", false, actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/docs/a.txt", false, 10, "", actor); err != nil {
		t.Fatal(err)
	}
	if err := Rename(diskPath, activePartitionStart(t), "/docs", "files", actor); err != nil {
		t.Fatalf("rename directory: %v", err)
	}
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sb := activeSuperBlock(t, file)
	if _, _, err := ResolvePath(file, sb, "/files/a.txt"); err != nil {
		t.Fatalf("renamed directory child missing: %v", err)
	}
}

func TestRenameRejectsRootDuplicateAndInvalidNames(t *testing.T) {
	resetMount(t)
	diskPath := setupFormattedFS(t)
	actor := rootActor()
	if err := Mkfile(diskPath, activePartitionStart(t), "/a.txt", false, 10, "", actor); err != nil {
		t.Fatal(err)
	}
	if err := Mkfile(diskPath, activePartitionStart(t), "/b.txt", false, 10, "", actor); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		path string
		name string
	}{
		{"/", "root2"},
		{"/a.txt", "b.txt"},
		{"/a.txt", "dir/name"},
		{"/a.txt", "nombre_muy_largo"},
		{"/missing.txt", "new.txt"},
	}
	for _, test := range cases {
		if err := Rename(diskPath, activePartitionStart(t), test.path, test.name, actor); err == nil {
			t.Fatalf("expected rename error for path=%q name=%q", test.path, test.name)
		}
	}
}

func TestRenameRequiresWritePermission(t *testing.T) {
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
	structs.SetPerm(&inode.IPerm, "444")
	if err := WriteInode(file, sb, index, inode); err != nil {
		file.Close()
		t.Fatal(err)
	}
	file.Close()

	err = Rename(diskPath, activePartitionStart(t), "/locked.txt", "other.txt", Actor{User: "user", UID: 2, GID: 2})
	if err == nil || !strings.Contains(err.Error(), "permiso") {
		t.Fatalf("expected permission error, got %v", err)
	}
}
