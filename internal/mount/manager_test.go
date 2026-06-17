package mount

import (
	"path/filepath"
	"testing"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/partition"
)

func TestMountIDsAndUnmount(t *testing.T) {
	manager := NewManager()
	dir := t.TempDir()
	disk1 := filepath.Join(dir, "d1.mia")
	disk2 := filepath.Join(dir, "d2.mia")

	createDiskWithPartition(t, disk1, "Part1")
	if err := partition.Create(partition.CreateOptions{Size: 100, Unit: "K", Path: disk1, Type: "P", Name: "Part2"}); err != nil {
		t.Fatalf("Create second partition failed: %v", err)
	}
	createDiskWithPartition(t, disk2, "Part1")

	first, err := manager.Mount(disk1, "Part1")
	if err != nil {
		t.Fatalf("Mount first failed: %v", err)
	}
	if first.ID != "961A" {
		t.Fatalf("first ID = %s, want 961A", first.ID)
	}

	second, err := manager.Mount(disk1, "Part2")
	if err != nil {
		t.Fatalf("Mount second failed: %v", err)
	}
	if second.ID != "961B" {
		t.Fatalf("second ID = %s, want 961B", second.ID)
	}

	third, err := manager.Mount(disk2, "Part1")
	if err != nil {
		t.Fatalf("Mount third failed: %v", err)
	}
	if third.ID != "962A" {
		t.Fatalf("third ID = %s, want 962A", third.ID)
	}

	if _, ok := manager.GetMounted("961a"); !ok {
		t.Fatal("expected case-insensitive lookup")
	}
	if err := manager.Unmount("961a"); err != nil {
		t.Fatalf("Unmount failed: %v", err)
	}
	if _, ok := manager.GetMounted("961A"); ok {
		t.Fatal("expected partition to be unmounted")
	}
	if err := manager.Unmount("nope"); err == nil {
		t.Fatal("expected invalid unmount error")
	}
}

func TestUnmountByDisk(t *testing.T) {
	manager := NewManager()
	path := filepath.Join(t.TempDir(), "d1.mia")
	createDiskWithPartition(t, path, "Part1")

	if _, err := manager.Mount(path, "Part1"); err != nil {
		t.Fatalf("Mount failed: %v", err)
	}
	manager.UnmountByDisk(path)
	if len(manager.List()) != 0 {
		t.Fatal("expected no mounts after UnmountByDisk")
	}
}

func createDiskWithPartition(t *testing.T, path string, name string) {
	t.Helper()
	if err := disk.MakeDisk(disk.MakeDiskOptions{Size: 1, Unit: "M", Path: path}); err != nil {
		t.Fatalf("MakeDisk failed: %v", err)
	}
	if err := partition.Create(partition.CreateOptions{Size: 100, Unit: "K", Path: path, Type: "P", Name: name}); err != nil {
		t.Fatalf("Create partition failed: %v", err)
	}
}
