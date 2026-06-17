package parser

import "testing"

func TestParseQuotedPath(t *testing.T) {
	cmd, skip, err := ParseLine(`mkdisk -size=10 -unit=M -path="/tmp/Disco 1.mia"`, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skip {
		t.Fatal("expected command, got skip")
	}
	if cmd.Name != "mkdisk" {
		t.Fatalf("expected mkdisk, got %q", cmd.Name)
	}
	if cmd.Params["path"] != "/tmp/Disco 1.mia" {
		t.Fatalf("unexpected path: %q", cmd.Params["path"])
	}
}

func TestParseCaseInsensitiveCommandAndParams(t *testing.T) {
	cmd, _, err := ParseLine(`MKDISK -Size=10 -Path=/tmp/a.mia`, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Name != "mkdisk" {
		t.Fatalf("expected mkdisk, got %q", cmd.Name)
	}
	if _, ok := cmd.Params["size"]; !ok {
		t.Fatal("expected normalized size param")
	}
	if cmd.Params["path"] != "/tmp/a.mia" {
		t.Fatalf("unexpected path: %q", cmd.Params["path"])
	}
}

func TestParseFlags(t *testing.T) {
	cmd, _, err := ParseLine(`mkfile -path=/home/a.txt -r -size=15`, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cmd.Flags["r"] {
		t.Fatal("expected -r flag")
	}
}

func TestParseComment(t *testing.T) {
	cmd, skip, err := ParseLine(`# esto es comentario`, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skip {
		t.Fatal("expected comment command, got skip")
	}
	if !cmd.IsComment() {
		t.Fatalf("expected comment command, got %q", cmd.Name)
	}
}

func TestParseEmptyLine(t *testing.T) {
	_, skip, err := ParseLine(`   `, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !skip {
		t.Fatal("expected skip")
	}
}

func TestParseUnclosedQuote(t *testing.T) {
	_, _, err := ParseLine(`mkdisk -path="/tmp/Disco 1.mia`, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAliases(t *testing.T) {
	cmd, _, err := ParseLine(`mkfile -s=15 -path=/home/a.txt`, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Params["size"] != "15" {
		t.Fatalf("expected size alias, got %#v", cmd.Params)
	}

	cmd, _, err = ParseLine(`Mkuser -usr=user1 -pass=123 -grupo=root`, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Name != "mkusr" {
		t.Fatalf("expected mkusr alias, got %q", cmd.Name)
	}
	if cmd.Params["user"] != "user1" || cmd.Params["grp"] != "root" {
		t.Fatalf("expected user/grp aliases, got %#v", cmd.Params)
	}
}

func TestBatTolerances(t *testing.T) {
	cmd, _, err := ParseLine(`mkfs -type-=full -id-=XXXX`, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Params["type"] != "full" || cmd.Params["id"] != "XXXX" {
		t.Fatalf("unexpected params: %#v", cmd.Params)
	}

	cmd, _, err = ParseLine(`mkusr -user=user1 -pass=123 -grp==root`, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Params["grp"] != "root" {
		t.Fatalf("expected grp=root, got %#v", cmd.Params)
	}
}

func TestValuesStayCaseSensitive(t *testing.T) {
	cmd, _, err := ParseLine(`login -user=Root -pass=MiPass -id=961A`, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Params["user"] != "Root" || cmd.Params["pass"] != "MiPass" || cmd.Params["id"] != "961A" {
		t.Fatalf("values should keep original case: %#v", cmd.Params)
	}
}
