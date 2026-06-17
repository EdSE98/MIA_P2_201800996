package structs

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestBinarySizes(t *testing.T) {
	tests := []struct {
		name string
		data any
		want int
	}{
		{"Content", Content{}, 16},
		{"FolderBlock", FolderBlock{}, 64},
		{"FileBlock", FileBlock{}, 64},
		{"PointerBlock", PointerBlock{}, 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := binary.Size(tt.data); got != tt.want {
				t.Fatalf("binary.Size(%s) = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

func TestPersistedStructSizesAreFixed(t *testing.T) {
	tests := []struct {
		name string
		data any
	}{
		{"MBR", MBR{}},
		{"Partition", Partition{}},
		{"EBR", EBR{}},
		{"SuperBlock", SuperBlock{}},
		{"Inode", Inode{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := binary.Size(tt.data); got <= 0 {
				t.Fatalf("binary.Size(%s) = %d, want > 0", tt.name, got)
			}
		})
	}
}

func TestNewEmptyInodeInitializesBlocks(t *testing.T) {
	inode := NewEmptyInode()
	for i, pointer := range inode.IBlock {
		if pointer != -1 {
			t.Fatalf("IBlock[%d] = %d, want -1", i, pointer)
		}
	}
}

func TestNewEmptyFolderBlockInitializesEntries(t *testing.T) {
	block := NewEmptyFolderBlock()
	for i, content := range block.BContent {
		if content.BInodo != -1 {
			t.Fatalf("BContent[%d].BInodo = %d, want -1", i, content.BInodo)
		}
	}
}

func TestNewEmptyPointerBlockInitializesPointers(t *testing.T) {
	block := NewEmptyPointerBlock()
	for i, pointer := range block.BPointers {
		if pointer != -1 {
			t.Fatalf("BPointers[%d] = %d, want -1", i, pointer)
		}
	}
}

func TestCopyStringFillsAndTruncates(t *testing.T) {
	dst := [5]byte{'x', 'x', 'x', 'x', 'x'}
	CopyString(dst[:], "abc")

	if got := string(dst[:3]); got != "abc" {
		t.Fatalf("copied value = %q, want abc", got)
	}
	if dst[3] != 0 || dst[4] != 0 {
		t.Fatalf("expected zero fill, got %#v", dst)
	}

	CopyString(dst[:], "abcdef")
	if got := string(dst[:]); got != "abcde" {
		t.Fatalf("truncated value = %q, want abcde", got)
	}
}

func TestFixedBytesToString(t *testing.T) {
	src := []byte{'r', 'o', 'o', 't', 0, 0}
	if got := FixedBytesToString(src); got != "root" {
		t.Fatalf("FixedBytesToString = %q, want root", got)
	}
}

func TestSetID4(t *testing.T) {
	var id [4]byte
	SetID4(&id, "961A")
	if got := string(id[:]); got != "961A" {
		t.Fatalf("id = %q, want 961A", got)
	}
}

func TestNewDateBytes(t *testing.T) {
	date := time.Date(2026, 6, 16, 12, 30, 45, 0, time.UTC)
	got := NewDateBytes(date)
	if FixedBytesToString(got[:]) != "2026-06-16 12:30:45" {
		t.Fatalf("unexpected date bytes: %q", FixedBytesToString(got[:]))
	}
}
