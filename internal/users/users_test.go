package users

import (
	"errors"
	"testing"
)

const initialContent = "1,G,root\n1,U,root,root,123\n"

func TestParseInitialUsersFile(t *testing.T) {
	groups, records, err := ParseUsersFile(initialContent)
	if err != nil {
		t.Fatalf("ParseUsersFile failed: %v", err)
	}
	if len(groups) != 1 || groups[0].ID != 1 || groups[0].Name != "root" || !groups[0].Active {
		t.Fatalf("unexpected groups: %#v", groups)
	}
	if len(records) != 1 || records[0].ID != 1 || records[0].Group != "root" || records[0].Username != "root" || records[0].Password != "123" {
		t.Fatalf("unexpected users: %#v", records)
	}
}

func TestFindRootUserAndGroup(t *testing.T) {
	user, ok, err := FindActiveUser(initialContent, "root")
	if err != nil {
		t.Fatalf("FindActiveUser failed: %v", err)
	}
	if !ok || user.Username != "root" {
		t.Fatalf("expected root user, got %#v ok=%v", user, ok)
	}

	group, ok, err := FindActiveGroup(initialContent, "root")
	if err != nil {
		t.Fatalf("FindActiveGroup failed: %v", err)
	}
	if !ok || group.ID != 1 {
		t.Fatalf("expected root group, got %#v ok=%v", group, ok)
	}

	gid, err := GroupIDForUser(initialContent, user)
	if err != nil {
		t.Fatalf("GroupIDForUser failed: %v", err)
	}
	if gid != 1 {
		t.Fatalf("gid = %d, want 1", gid)
	}
}

func TestInactiveRecordsAreIgnored(t *testing.T) {
	content := "0,G,old\n1,G,root\n0,U,root,old,old\n1,U,root,root,123\n"
	if _, ok, err := FindActiveGroup(content, "old"); err != nil || ok {
		t.Fatalf("expected inactive group ignored, ok=%v err=%v", ok, err)
	}
	if _, ok, err := FindActiveUser(content, "old"); err != nil || ok {
		t.Fatalf("expected inactive user ignored, ok=%v err=%v", ok, err)
	}
}

func TestMalformedLineReturnsError(t *testing.T) {
	if _, _, err := ParseUsersFile("1,U,root\n"); err == nil {
		t.Fatal("expected malformed line error")
	}
}

func TestUserAndPasswordAreCaseSensitive(t *testing.T) {
	if _, ok, err := FindActiveUser(initialContent, "Root"); err != nil || ok {
		t.Fatalf("Root should not match root, ok=%v err=%v", ok, err)
	}

	_, _, err := Authenticate(initialContent, "root", "123")
	if err != nil {
		t.Fatalf("Authenticate root failed: %v", err)
	}
	_, _, err = Authenticate(initialContent, "root", "1234")
	if !errors.Is(err, ErrBadPassword) {
		t.Fatalf("expected bad password, got %v", err)
	}
}
