package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListDisksUsesConfiguredDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MIA_DISKS_DIR", dir)

	if err := os.WriteFile(filepath.Join(dir, "d1.dsk"), []byte("disk"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "d2.mia"), []byte("mia"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nota.txt"), []byte("skip"), 0o644); err != nil {
		t.Fatal(err)
	}

	disks, err := ListDisks()
	if err != nil {
		t.Fatal(err)
	}
	if len(disks) != 2 {
		t.Fatalf("expected 2 disks, got %d", len(disks))
	}
	if disks[0].Name != "d1.dsk" || disks[1].Name != "d2.mia" {
		t.Fatalf("unexpected disks: %+v", disks)
	}
}
