package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/session"
)

func TestExecuteCommandCreatesDiskAndCapturesOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "console.dsk")
	result, err := ExecuteCommand(`mkdisk -size=1 -unit=M -path="` + path + `"`)
	if err != nil {
		t.Fatal(err)
	}
	if result.Command == "" || !strings.Contains(result.Output, "Disco creado:") {
		t.Fatalf("unexpected result: %+v", result)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("disk was not created: %v", err)
	}
}

func TestExecuteCommandRejectsUnsafeInput(t *testing.T) {
	for _, input := range []string{
		"",
		"   ",
		"# comentario",
		"pause",
		"exit",
		"mkdir -path=/a\nlogout",
		"mkdir -path=/a\rlogout",
	} {
		if _, err := ExecuteCommand(input); err == nil {
			t.Fatalf("expected error for %q", input)
		}
	}
}

func TestExecuteCommandUsesGlobalSessionAndDispatcher(t *testing.T) {
	diskPath := setupFSOperationsService(t)
	result, err := ExecuteCommand("mkdir -p -path=/console/docs")
	if err != nil {
		t.Fatal(err)
	}
	if result.Session == nil || result.Session.User != "root" {
		t.Fatalf("active session missing from result: %+v", result)
	}
	if !strings.Contains(result.Output, "Carpeta creada: /console/docs") {
		t.Fatalf("unexpected output: %q", result.Output)
	}

	active, err := session.RequireActive()
	if err != nil {
		t.Fatal(err)
	}
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sb, err := fs.ReadSuperBlock(file, int64(active.PartitionStart))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := fs.ResolvePath(file, sb, "/console/docs"); err != nil {
		t.Fatalf("dispatcher did not persist directory: %v", err)
	}
}
