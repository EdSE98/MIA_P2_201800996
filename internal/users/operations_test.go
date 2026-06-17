package users

import (
	"strings"
	"testing"
)

func TestMakeGroup(t *testing.T) {
	content, err := MakeGroup(initialContent, "usuarios")
	if err != nil {
		t.Fatalf("MakeGroup failed: %v", err)
	}
	if !strings.Contains(content, "2,G,usuarios\n") {
		t.Fatalf("expected group with incremental ID, got %q", content)
	}
	if _, err := MakeGroup(content, "usuarios"); err == nil {
		t.Fatal("expected duplicate group error")
	}
	if _, err := MakeGroup(content, "larguisimo11"); err == nil {
		t.Fatal("expected long group error")
	}
	if content, err = MakeGroup(content, "Usuarios"); err != nil {
		t.Fatalf("case-different group should be allowed: %v", err)
	}
	if !strings.Contains(content, "3,G,Usuarios\n") {
		t.Fatalf("expected case-sensitive group, got %q", content)
	}
}

func TestRemoveGroup(t *testing.T) {
	content, err := MakeGroup(initialContent, "usuarios")
	if err != nil {
		t.Fatalf("MakeGroup failed: %v", err)
	}
	content, err = RemoveGroup(content, "usuarios")
	if err != nil {
		t.Fatalf("RemoveGroup failed: %v", err)
	}
	if !strings.Contains(content, "0,G,usuarios\n") {
		t.Fatalf("expected removed group preserved, got %q", content)
	}
	if _, err := RemoveGroup(content, "missing"); err == nil {
		t.Fatal("expected missing group error")
	}
	if _, err := RemoveGroup(content, "root"); err == nil {
		t.Fatal("expected root group removal error")
	}
	content, err = MakeGroup(content, "nuevo")
	if err != nil {
		t.Fatalf("MakeGroup after removal failed: %v", err)
	}
	if !strings.Contains(content, "3,G,nuevo\n") {
		t.Fatalf("expected IDs not reused, got %q", content)
	}
}

func TestMakeUser(t *testing.T) {
	content, err := MakeGroup(initialContent, "usuarios")
	if err != nil {
		t.Fatalf("MakeGroup failed: %v", err)
	}
	content, err = MakeUser(content, "user1", "usuario", "usuarios")
	if err != nil {
		t.Fatalf("MakeUser failed: %v", err)
	}
	if !strings.Contains(content, "2,U,usuarios,user1,usuario\n") {
		t.Fatalf("expected user with incremental ID, got %q", content)
	}
	if _, err := MakeUser(content, "user1", "usuario", "usuarios"); err == nil {
		t.Fatal("expected duplicate user error")
	}
	if _, err := MakeUser(content, "larguisimo11", "usuario", "usuarios"); err == nil {
		t.Fatal("expected long username error")
	}
	if _, err := MakeUser(content, "user2", "larguisimo11", "usuarios"); err == nil {
		t.Fatal("expected long password error")
	}
	if _, err := MakeUser(content, "user2", "usuario", "inexistente"); err == nil {
		t.Fatal("expected missing group error")
	}
}

func TestRemoveUser(t *testing.T) {
	content, err := MakeGroup(initialContent, "usuarios")
	if err != nil {
		t.Fatalf("MakeGroup failed: %v", err)
	}
	content, err = MakeUser(content, "user1", "usuario", "usuarios")
	if err != nil {
		t.Fatalf("MakeUser failed: %v", err)
	}
	content, err = RemoveUser(content, "user1")
	if err != nil {
		t.Fatalf("RemoveUser failed: %v", err)
	}
	if !strings.Contains(content, "0,U,usuarios,user1,usuario\n") {
		t.Fatalf("expected removed user preserved, got %q", content)
	}
	if _, err := RemoveUser(content, "missing"); err == nil {
		t.Fatal("expected missing user error")
	}
	if _, err := RemoveUser(content, "root"); err == nil {
		t.Fatal("expected root user removal error")
	}
	content, err = MakeUser(content, "user2", "usuario", "usuarios")
	if err != nil {
		t.Fatalf("MakeUser after removal failed: %v", err)
	}
	if !strings.Contains(content, "3,U,usuarios,user2,usuario\n") {
		t.Fatalf("expected IDs not reused, got %q", content)
	}
}

func TestChangeUserGroup(t *testing.T) {
	content, err := MakeGroup(initialContent, "usuarios")
	if err != nil {
		t.Fatalf("MakeGroup failed: %v", err)
	}
	content, err = MakeUser(content, "user1", "usuario", "usuarios")
	if err != nil {
		t.Fatalf("MakeUser failed: %v", err)
	}
	content, err = ChangeUserGroup(content, "user1", "root")
	if err != nil {
		t.Fatalf("ChangeUserGroup failed: %v", err)
	}
	if !strings.Contains(content, "2,U,root,user1,usuario\n") {
		t.Fatalf("expected changed group, got %q", content)
	}
	if _, err := ChangeUserGroup(content, "missing", "root"); err == nil {
		t.Fatal("expected missing user error")
	}
	if _, err := ChangeUserGroup(content, "user1", "missing"); err == nil {
		t.Fatal("expected missing group error")
	}
}

func TestSerializeUsersFileEndsWithNewline(t *testing.T) {
	records, err := ParseUsersRecords("0,G,old\n1,G,root\n")
	if err != nil {
		t.Fatalf("ParseUsersRecords failed: %v", err)
	}
	content := SerializeUsersFile(records)
	if content != "0,G,old\n1,G,root\n" {
		t.Fatalf("unexpected serialization: %q", content)
	}
}
