package fs

import (
	"testing"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/mount"
)

func TestCleanAbsPath(t *testing.T) {
	parts, err := CleanAbsPath("/")
	if err != nil {
		t.Fatalf("CleanAbsPath / failed: %v", err)
	}
	if len(parts) != 0 {
		t.Fatalf("expected empty root parts, got %#v", parts)
	}

	parts, err = CleanAbsPath("/home//archivos")
	if err != nil {
		t.Fatalf("CleanAbsPath failed: %v", err)
	}
	if len(parts) != 2 || parts[0] != "home" || parts[1] != "archivos" {
		t.Fatalf("unexpected parts: %#v", parts)
	}

	if _, err := CleanAbsPath("home/archivos"); err == nil {
		t.Fatal("expected relative path error")
	}
	if _, err := CleanAbsPath("/nombre_demasiado_largo"); err == nil {
		t.Fatal("expected long name error")
	}
	if _, err := CleanAbsPath("/home/../x"); err == nil {
		t.Fatal("expected dotdot error")
	}
}

func TestResolveRootAndUsers(t *testing.T) {
	resetMount(t)
	path := createMountedPartition(t, 10, 5, "P")
	var noop testWriter
	if err := FormatFromParams(map[string]string{"id": "961A"}, noop); err != nil {
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

	index, inode, err := ResolvePath(file, sb, "/")
	if err != nil {
		t.Fatalf("ResolvePath root failed: %v", err)
	}
	if index != 0 || inode.IType != '0' {
		t.Fatalf("unexpected root: index=%d inode=%#v", index, inode)
	}

	index, inode, err = ResolvePath(file, sb, "/users.txt")
	if err != nil {
		t.Fatalf("ResolvePath users failed: %v", err)
	}
	if index != 1 || inode.IType != '1' {
		t.Fatalf("unexpected users.txt: index=%d inode=%#v", index, inode)
	}
}

type testWriter struct{}

func (testWriter) Write(p []byte) (int, error) { return len(p), nil }
