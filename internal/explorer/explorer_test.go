package explorer

import (
	"path/filepath"
	"strings"
	"testing"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/partition"
)

func TestExplorerListReadAndStat(t *testing.T) {
	id := setupExplorerDisk(t)

	listing, err := List(id, "/")
	if err != nil {
		t.Fatal(err)
	}
	if listing.ID != id || listing.Path != "/" {
		t.Fatalf("unexpected listing metadata: %+v", listing)
	}
	if len(listing.Items) != 1 {
		t.Fatalf("expected users.txt in root, got %+v", listing.Items)
	}
	item := listing.Items[0]
	if item.Name != "users.txt" || item.Path != "/users.txt" || item.Type != "file" {
		t.Fatalf("unexpected root item: %+v", item)
	}
	if item.Owner != "root" || item.Group != "root" || item.Permissions != "664" {
		t.Fatalf("unexpected identity/permissions: %+v", item)
	}

	content, err := Read(id, "/users.txt")
	if err != nil {
		t.Fatal(err)
	}
	if content.Name != "users.txt" || !strings.Contains(content.Content, "1,U,root,root,123") {
		t.Fatalf("unexpected content: %+v", content)
	}

	stat, err := Stat(id, "/users.txt")
	if err != nil {
		t.Fatal(err)
	}
	if stat.Inode != 1 || stat.Type != "file" || stat.Size != content.Size {
		t.Fatalf("unexpected stat: %+v", stat)
	}
}

func TestExplorerTypeErrors(t *testing.T) {
	id := setupExplorerDisk(t)

	if _, err := List(id, "/users.txt"); err == nil {
		t.Fatal("expected list file error")
	}
	if _, err := Read(id, "/"); err == nil {
		t.Fatal("expected read directory error")
	}
}

func setupExplorerDisk(t *testing.T) string {
	t.Helper()
	mount.Global = mount.NewManager()

	path := filepath.Join(t.TempDir(), "explorer.dsk")
	if err := disk.MakeDisk(disk.MakeDiskOptions{Size: 10, Unit: "M", Fit: "FF", Path: path}); err != nil {
		t.Fatal(err)
	}
	if err := partition.Create(partition.CreateOptions{Size: 5, Unit: "M", Path: path, Type: "P", Fit: "FF", Name: "Part1"}); err != nil {
		t.Fatal(err)
	}
	mounted, err := mount.Global.Mount(path, "Part1")
	if err != nil {
		t.Fatal(err)
	}
	if err := fs.Format(fs.FormatOptions{ID: mounted.ID, Type: "full"}, discardWriter{}); err != nil {
		t.Fatal(err)
	}
	return mounted.ID
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
