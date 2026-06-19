package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mia_p1_201800996/internal/commands"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/session"
)

func TestScriptRunsDiskReports(t *testing.T) {
	resetCLIManagers(t)

	dir := t.TempDir()
	diskPath := filepath.Join(dir, "script.mia")
	mbrReport := filepath.Join(dir, "mbr.dot")
	diskReport := filepath.Join(dir, "disk.dot")
	scriptPath := filepath.Join(dir, "input.smia")
	script := strings.Join([]string{
		"mkdisk -size=2 -unit=M -path=\"" + diskPath + "\"",
		"fdisk -size=200 -unit=K -type=P -name=Part1 -path=\"" + diskPath + "\"",
		"mount -path=\"" + diskPath + "\" -name=Part1",
		"rep -name=mbr -path=\"" + mbrReport + "\" -id=961A",
		"rep -name=disk -path=\"" + diskReport + "\" -id=961A",
	}, "\n")

	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var out bytes.Buffer
	dispatcher := commands.NewDispatcher(strings.NewReader(""), &out)
	runner := New(dispatcher, strings.NewReader(""), &out)
	if err := runner.RunScript(scriptPath); err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Particiones montadas:") || !strings.Contains(output, "961A") || !strings.Contains(output, "Part1") {
		t.Fatalf("expected mounted partition listing, got:\n%s", output)
	}

	if _, err := os.Stat(mbrReport); err != nil {
		t.Fatalf("expected mbr report: %v\noutput:\n%s", err, out.String())
	}
	if _, err := os.Stat(diskReport); err != nil {
		t.Fatalf("expected disk report: %v\noutput:\n%s", err, out.String())
	}
}

func TestScriptRunsLoginLogout(t *testing.T) {
	resetCLIManagers(t)
	dir := t.TempDir()
	diskPath := filepath.Join(dir, "session.mia")
	scriptPath := filepath.Join(dir, "session.smia")
	script := strings.Join([]string{
		"mkdisk -size=10 -unit=M -path=\"" + diskPath + "\"",
		"fdisk -size=5 -unit=M -path=\"" + diskPath + "\" -name=Part1",
		"mount -path=\"" + diskPath + "\" -name=Part1",
		"mkfs -id=961A",
		"login -user=root -pass=123 -id=961A",
		"logout",
	}, "\n")
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var out bytes.Buffer
	dispatcher := commands.NewDispatcher(strings.NewReader(""), &out)
	runner := New(dispatcher, strings.NewReader(""), &out)
	if err := runner.RunScript(scriptPath); err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}
	if !strings.Contains(out.String(), "Sesion iniciada: root en 961A") {
		t.Fatalf("expected login success, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Sesion cerrada") {
		t.Fatalf("expected logout success, got:\n%s", out.String())
	}
}

func TestPendingFilesystemCommandsRequireSession(t *testing.T) {
	resetCLIManagers(t)
	dir := t.TempDir()
	diskPath := filepath.Join(dir, "session-required.mia")
	scriptPath := filepath.Join(dir, "session-required.smia")
	script := strings.Join([]string{
		"mkgrp -name=usuarios",
		"mkdisk -size=10 -unit=M -path=\"" + diskPath + "\"",
		"fdisk -size=5 -unit=M -path=\"" + diskPath + "\" -name=Part1",
		"mount -path=\"" + diskPath + "\" -name=Part1",
		"mkfs -id=961A",
		"login -user=root -pass=123 -id=961A",
		"mkgrp -name=usuarios",
	}, "\n")
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var out bytes.Buffer
	dispatcher := commands.NewDispatcher(strings.NewReader(""), &out)
	runner := New(dispatcher, strings.NewReader(""), &out)
	if err := runner.RunScript(scriptPath); err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "necesita iniciar sesion") {
		t.Fatalf("expected session required error, got:\n%s", output)
	}
	if !strings.Contains(output, "Grupo creado: usuarios") {
		t.Fatalf("expected mkgrp success after login, got:\n%s", output)
	}
}

func TestScriptRunsUserAdministrationFlow(t *testing.T) {
	resetCLIManagers(t)
	dir := t.TempDir()
	diskPath := filepath.Join(dir, "users.mia")
	scriptPath := filepath.Join(dir, "users.smia")
	script := strings.Join([]string{
		"mkdisk -size=10 -unit=M -path=\"" + diskPath + "\"",
		"fdisk -size=5 -unit=M -path=\"" + diskPath + "\" -name=Part1",
		"mount -path=\"" + diskPath + "\" -name=Part1",
		"mkfs -id=961A",
		"login -user=root -pass=123 -id=961A",
		"mkgrp -name=usuarios",
		"mkusr -user=user1 -pass=usuario -grp=usuarios",
		"logout",
		"login -user=user1 -pass=usuario -id=961A",
		"mkgrp -name=fallo",
		"logout",
		"login -user=root -pass=123 -id=961A",
		"chgrp -user=user1 -grp=root",
		"rmusr -user=user1",
		"rmgrp -name=usuarios",
		"logout",
		"login -user=user1 -pass=usuario -id=961A",
	}, "\n")
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var out bytes.Buffer
	dispatcher := commands.NewDispatcher(strings.NewReader(""), &out)
	runner := New(dispatcher, strings.NewReader(""), &out)
	if err := runner.RunScript(scriptPath); err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}
	output := out.String()
	for _, want := range []string{
		"Grupo creado: usuarios",
		"Usuario creado: user1",
		"Sesion iniciada: user1 en 961A",
		"solo el usuario root puede ejecutar este comando",
		"Grupo de usuario actualizado: user1 -> root",
		"Usuario eliminado: user1",
		"Grupo eliminado: usuarios",
		"el usuario no existe",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output:\n%s", want, output)
		}
	}
}

func resetCLIManagers(t *testing.T) {
	t.Helper()
	oldMount := mount.Global
	oldSession := session.Global
	mount.Global = mount.NewManager()
	session.Global = session.NewManager()
	t.Cleanup(func() {
		mount.Global = oldMount
		session.Global = oldSession
	})
}
