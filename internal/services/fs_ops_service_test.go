package services

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/partition"
	"mia_p1_201800996/internal/session"
)

func TestFSOperationsRequireActiveSession(t *testing.T) {
	oldSession := session.Global
	session.Global = session.NewManager()
	t.Cleanup(func() { session.Global = oldSession })

	if err := EditFile(dto.EditFileRequest{Path: "/a.txt", Contenido: "/tmp/a.txt"}); err == nil {
		t.Fatal("expected edit session error")
	}
	if err := RenameEntry(dto.RenameEntryRequest{Path: "/a.txt", Name: "b.txt"}); err == nil {
		t.Fatal("expected rename session error")
	}
	if err := RemoveEntry(dto.RemoveEntryRequest{Path: "/a.txt"}); err == nil {
		t.Fatal("expected remove session error")
	}
}

func TestEditAndRenameServicesUseActiveSession(t *testing.T) {
	diskPath := setupFSOperationsService(t)
	contentPath := filepath.Join(t.TempDir(), "new.txt")
	if err := os.WriteFile(contentPath, []byte("contenido editado"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := EditFile(dto.EditFileRequest{
		Path: "/home/docs/a.txt", Contenido: contentPath,
	}); err != nil {
		t.Fatalf("EditFile: %v", err)
	}
	if err := RenameEntry(dto.RenameEntryRequest{
		Path: "/home/docs/a.txt", Name: "b1.txt",
	}); err != nil {
		t.Fatalf("RenameEntry: %v", err)
	}

	active, err := session.RequireActive()
	if err != nil {
		t.Fatal(err)
	}
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	sb, err := fs.ReadSuperBlock(file, int64(active.PartitionStart))
	if err != nil {
		t.Fatal(err)
	}
	_, inode, err := fs.ResolvePath(file, sb, "/home/docs/b1.txt")
	if err != nil {
		t.Fatal(err)
	}
	content, err := fs.ReadFileContent(file, sb, inode)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "contenido editado" {
		t.Fatalf("content = %q", content)
	}
	file.Close()

	if err := RemoveEntry(dto.RemoveEntryRequest{Path: "/home/docs/b1.txt"}); err != nil {
		t.Fatalf("RemoveEntry: %v", err)
	}
	file, _, err = disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sb, err = fs.ReadSuperBlock(file, int64(active.PartitionStart))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := fs.ResolvePath(file, sb, "/home/docs/b1.txt"); err == nil {
		t.Fatal("removed entry still exists")
	}
}

func setupFSOperationsService(t *testing.T) string {
	t.Helper()
	oldMount := mount.Global
	oldSession := session.Global
	mount.Global = mount.NewManager()
	session.Global = session.NewManager()
	t.Cleanup(func() {
		mount.Global = oldMount
		session.Global = oldSession
	})

	path := filepath.Join(t.TempDir(), "fs-ops-service.dsk")
	if err := disk.MakeDisk(disk.MakeDiskOptions{Size: 10, Unit: "M", Path: path}); err != nil {
		t.Fatal(err)
	}
	if err := partition.Create(partition.CreateOptions{
		Size: 8, Unit: "M", Path: path, Type: "P", Name: "Part1",
	}); err != nil {
		t.Fatal(err)
	}
	mounted, err := mount.Global.Mount(path, "Part1")
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := fs.Format(fs.FormatOptions{ID: mounted.ID, Type: "full"}, &out); err != nil {
		t.Fatal(err)
	}
	active, err := session.Login("root", "123", mounted.ID)
	if err != nil {
		t.Fatal(err)
	}
	actor := fs.Actor{User: active.User, UID: active.UID, GID: active.GID}
	if err := fs.Mkdir(path, int64(active.PartitionStart), "/home/docs", true, actor); err != nil {
		t.Fatal(err)
	}
	if err := fs.Mkfile(path, int64(active.PartitionStart), "/home/docs/a.txt", false, 20, "", actor); err != nil {
		t.Fatal(err)
	}
	return path
}
