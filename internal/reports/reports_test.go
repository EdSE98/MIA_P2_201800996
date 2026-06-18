package reports

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/partition"
)

func TestGenerateMBRReportDot(t *testing.T) {
	manager := resetGlobalMount(t)
	_ = manager
	diskPath := createReportDisk(t)

	mounted, err := mount.Global.Mount(diskPath, "Primaria1")
	if err != nil {
		t.Fatalf("Mount failed: %v", err)
	}
	output := filepath.Join(t.TempDir(), "mbr.dot")

	var out bytes.Buffer
	if err := Generate(map[string]string{"name": "mbr", "path": output, "id": mounted.ID}, &out); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := readFile(t, output)
	for _, want := range []string{"mbr_tamano", "mbr_fecha_creacion", "mbr_dsk_signature", "part_name", "Primaria1", "Extendida1", "EBR"} {
		if !strings.Contains(content, want) {
			t.Fatalf("DOT does not contain %q:\n%s", want, content)
		}
	}
}

func TestGenerateDiskReportDotDoesNotGrowDisk(t *testing.T) {
	resetGlobalMount(t)
	diskPath := createReportDisk(t)
	before := fileSize(t, diskPath)

	mounted, err := mount.Global.Mount(diskPath, "Primaria1")
	if err != nil {
		t.Fatalf("Mount failed: %v", err)
	}
	output := filepath.Join(t.TempDir(), "disk.dot")

	var out bytes.Buffer
	if err := Generate(map[string]string{"name": "disk", "path": output, "id": mounted.ID}, &out); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := readFile(t, output)
	for _, want := range []string{"MBR", "Libre", "Primaria", "Extendida", "%"} {
		if !strings.Contains(content, want) {
			t.Fatalf("DOT does not contain %q:\n%s", want, content)
		}
	}
	after := fileSize(t, diskPath)
	if after != before {
		t.Fatalf("disk grew from %d to %d", before, after)
	}
}

func TestGenerateUnknownReportFails(t *testing.T) {
	resetGlobalMount(t)
	var out bytes.Buffer
	if err := Generate(map[string]string{"name": "desconocido", "path": "/tmp/x.dot", "id": "961A"}, &out); err == nil {
		t.Fatal("expected unknown report error")
	}
}

func createReportDisk(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "report.mia")
	if err := disk.MakeDisk(disk.MakeDiskOptions{Size: 2, Unit: "M", Path: path}); err != nil {
		t.Fatalf("MakeDisk failed: %v", err)
	}
	if err := partition.Create(partition.CreateOptions{Size: 200, Unit: "K", Path: path, Type: "P", Name: "Primaria1"}); err != nil {
		t.Fatalf("Create primary failed: %v", err)
	}
	if err := partition.Create(partition.CreateOptions{Size: 500, Unit: "K", Path: path, Type: "E", Name: "Extendida1"}); err != nil {
		t.Fatalf("Create extended failed: %v", err)
	}
	return path
}

func resetGlobalMount(t *testing.T) *mount.Manager {
	t.Helper()
	old := mount.Global
	manager := mount.NewManager()
	mount.Global = manager
	t.Cleanup(func() {
		mount.Global = old
	})
	return manager
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	return string(content)
}

func fileSize(t *testing.T, path string) int64 {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer file.Close()
	size, err := binio.FileSize(file)
	if err != nil {
		t.Fatalf("FileSize failed: %v", err)
	}
	return size
}
