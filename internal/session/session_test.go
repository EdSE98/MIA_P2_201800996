package session

import (
	"bytes"
	"strings"
	"testing"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/partition"
)

func TestLoginLogoutRoot(t *testing.T) {
	resetManagers(t)
	createFormattedPartition(t)

	logged, err := Login("root", "123", "961A")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if !logged.Active || logged.User != "root" || logged.UID != 1 || logged.Group != "root" || logged.GID != 1 || logged.MountedID != "961A" {
		t.Fatalf("unexpected session: %#v", logged)
	}

	current, ok := Current()
	if !ok || current.User != "root" {
		t.Fatalf("expected active root session, got %#v ok=%v", current, ok)
	}

	if _, err := Login("root", "123", "961A"); err == nil || !strings.Contains(err.Error(), "ya existe una sesion activa") {
		t.Fatalf("expected duplicate login error, got %v", err)
	}
	if err := Logout(); err != nil {
		t.Fatalf("Logout failed: %v", err)
	}
	if _, ok := Current(); ok {
		t.Fatal("expected session cleared")
	}
	if err := Logout(); err == nil || !strings.Contains(err.Error(), "no existe sesion activa") {
		t.Fatalf("expected logout without session error, got %v", err)
	}
}

func TestLoginFailures(t *testing.T) {
	resetManagers(t)
	createFormattedPartition(t)

	if _, err := Login("roca", "123", "961A"); err == nil || !strings.Contains(err.Error(), "el usuario no existe") {
		t.Fatalf("expected missing user error, got %v", err)
	}
	if _, err := Login("root", "bad", "961A"); err == nil || !strings.Contains(err.Error(), "autenticacion fallida") {
		t.Fatalf("expected bad password error, got %v", err)
	}
	if _, err := Login("root", "123", "999A"); err == nil || !strings.Contains(err.Error(), "no existe montaje") {
		t.Fatalf("expected invalid id error, got %v", err)
	}
	if _, err := Login("Root", "123", "961A"); err == nil || !strings.Contains(err.Error(), "el usuario no existe") {
		t.Fatalf("expected case-sensitive user error, got %v", err)
	}
	if _, err := Login("root", "123", "961a"); err != nil {
		t.Fatalf("expected case-insensitive id login, got %v", err)
	}
}

func TestLoginRejectsUnformattedPartition(t *testing.T) {
	resetManagers(t)
	path := createMountedPartition(t)

	_ = path
	if _, err := Login("root", "123", "961A"); err == nil || !strings.Contains(err.Error(), "no esta formateada") {
		t.Fatalf("expected unformatted partition error, got %v", err)
	}
}

func TestClearIfMountedID(t *testing.T) {
	resetManagers(t)
	createFormattedPartition(t)

	if _, err := Login("root", "123", "961A"); err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if !ClearIfMountedID("961a") {
		t.Fatal("expected active session to be cleared")
	}
	if _, ok := Current(); ok {
		t.Fatal("expected no active session")
	}
}

func TestClearIfDiskPath(t *testing.T) {
	resetManagers(t)
	path := createFormattedPartition(t)

	if _, err := Login("root", "123", "961A"); err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if !ClearIfDiskPath(path) {
		t.Fatal("expected active session to be cleared by disk path")
	}
	if _, ok := Current(); ok {
		t.Fatal("expected no active session")
	}
}

func createFormattedPartition(t *testing.T) string {
	t.Helper()
	path := createMountedPartition(t)
	var out bytes.Buffer
	if err := fs.Format(fs.FormatOptions{ID: "961A", Type: "full"}, &out); err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	return path
}

func createMountedPartition(t *testing.T) string {
	t.Helper()
	path := t.TempDir() + "/session.mia"
	if err := disk.MakeDisk(disk.MakeDiskOptions{Size: 10, Unit: "M", Path: path}); err != nil {
		t.Fatalf("MakeDisk failed: %v", err)
	}
	if err := partition.Create(partition.CreateOptions{Size: 5, Unit: "M", Path: path, Type: "P", Name: "Part1"}); err != nil {
		t.Fatalf("Create partition failed: %v", err)
	}
	if _, err := mount.Global.Mount(path, "Part1"); err != nil {
		t.Fatalf("Mount failed: %v", err)
	}
	return path
}

func resetManagers(t *testing.T) {
	t.Helper()
	oldSession := Global
	oldMount := mount.Global
	Global = NewManager()
	mount.Global = mount.NewManager()
	t.Cleanup(func() {
		Global = oldSession
		mount.Global = oldMount
	})
}
