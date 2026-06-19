package reports

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/graphviz"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/partition"
	"mia_p1_201800996/internal/session"
)

func TestEXT2Reports(t *testing.T) {
	diskPath := setupEXT2ReportFS(t)
	before := fileSize(t, diskPath)
	dir := t.TempDir()

	cases := []struct {
		name     string
		path     string
		extra    map[string]string
		contains []string
	}{
		{"sb", filepath.Join(dir, "sb.dot"), nil, []string{"s_filesystem_type", "s_magic", "s_inode_start"}},
		{"inode", filepath.Join(dir, "inode.dot"), nil, []string{"Inodo 0", "Inodo 1", "i_block"}},
		{"block", filepath.Join(dir, "block.dot"), nil, []string{"Bloque", "users.txt", "0123456789", "PointerBlock"}},
		{"bm_inode", filepath.Join(dir, "bm_inode.txt"), nil, []string{"1 1", "\n"}},
		{"bm_block", filepath.Join(dir, "bm_block.txt"), nil, []string{"1 1", "\n"}},
		{"bm_bloc", filepath.Join(dir, "bm_bloc.txt"), nil, []string{"1 1", "\n"}},
		{"file", filepath.Join(dir, "file_a.txt"), map[string]string{"path_file_ls": "/home/docs/a.txt"}, []string{"Archivo: /home/docs/a.txt", "0123456789"}},
		{"ls", filepath.Join(dir, "ls_root.dot"), map[string]string{"path_file_ls": "/"}, []string{"users.txt", "home"}},
		{"ls", filepath.Join(dir, "ls_docs.dot"), map[string]string{"path_file_ls": "/home/docs"}, []string{"a.txt", "b.txt"}},
		{"tree", filepath.Join(dir, "tree.dot"), nil, []string{"Inodo 0", "users.txt", "home", "PointerBlock", "->"}},
	}

	for _, tt := range cases {
		t.Run(tt.name+"_"+filepath.Base(tt.path), func(t *testing.T) {
			params := map[string]string{"name": tt.name, "path": tt.path, "id": "961A"}
			for k, v := range tt.extra {
				params[k] = v
			}
			var out bytes.Buffer
			if err := Generate(params, &out); err != nil {
				t.Fatalf("Generate failed: %v\nout=%s", err, out.String())
			}
			content := readFile(t, tt.path)
			for _, want := range tt.contains {
				if !strings.Contains(content, want) {
					t.Fatalf("%s does not contain %q:\n%s", tt.name, want, content)
				}
			}
			if strings.Contains(content, "\x00") {
				t.Fatalf("report contains NUL bytes")
			}
		})
	}

	after := fileSize(t, diskPath)
	if after != before {
		t.Fatalf("disk changed size from %d to %d", before, after)
	}
}

func TestBitmapReportsRenderVisualFormats(t *testing.T) {
	setupEXT2ReportFS(t)
	dir := t.TempDir()
	for _, name := range []string{"bm_inode", "bm_block"} {
		t.Run(name, func(t *testing.T) {
			output := filepath.Join(dir, name+".pdf")
			var out bytes.Buffer
			if err := Generate(map[string]string{"name": name, "path": output, "id": "961A"}, &out); err != nil {
				t.Fatalf("Generate failed: %v", err)
			}
			dotPath := filepath.Join(dir, name+".dot")
			if _, err := os.Stat(dotPath); err != nil {
				t.Fatalf("expected dot fallback/auxiliary file: %v\nout=%s", err, out.String())
			}
			dot := readFile(t, dotPath)
			if !strings.Contains(dot, name) || !strings.Contains(dot, "1 1") {
				t.Fatalf("unexpected bitmap dot:\n%s", dot)
			}
			if graphviz.IsDotAvailable() {
				if _, err := os.Stat(output); err != nil {
					t.Fatalf("expected rendered pdf: %v\nout=%s", err, out.String())
				}
			} else if !strings.Contains(out.String(), "Advertencia") {
				t.Fatalf("expected Graphviz warning, got:\n%s", out.String())
			}
		})
	}
}

func TestEXT2ReportErrors(t *testing.T) {
	setupEXT2ReportFS(t)
	var out bytes.Buffer
	if err := Generate(map[string]string{"name": "sb", "path": filepath.Join(t.TempDir(), "sb.dot"), "id": "999A"}, &out); err == nil {
		t.Fatal("expected invalid id error")
	}
	if err := Generate(map[string]string{"name": "file", "path": filepath.Join(t.TempDir(), "file.txt"), "id": "961A", "path_file_ls": "/home/docs"}, &out); err == nil {
		t.Fatal("expected file report on directory error")
	}
	if err := Generate(map[string]string{"name": "file", "path": filepath.Join(t.TempDir(), "file.txt"), "id": "961A"}, &out); err == nil {
		t.Fatal("expected missing path_file_ls error")
	}
}

func TestEXT2ReportRejectsUnformattedPartition(t *testing.T) {
	resetReportManagers(t)
	path := filepath.Join(t.TempDir(), "unformatted.mia")
	if err := disk.MakeDisk(disk.MakeDiskOptions{Size: 10, Unit: "M", Path: path}); err != nil {
		t.Fatalf("MakeDisk failed: %v", err)
	}
	if err := partition.Create(partition.CreateOptions{Size: 5, Unit: "M", Path: path, Type: "P", Name: "Part1"}); err != nil {
		t.Fatalf("partition failed: %v", err)
	}
	if _, err := mount.Global.Mount(path, "Part1"); err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	var out bytes.Buffer
	if err := Generate(map[string]string{"name": "sb", "path": filepath.Join(t.TempDir(), "sb.dot"), "id": "961A"}, &out); err == nil {
		t.Fatal("expected unformatted error")
	}
}

func setupEXT2ReportFS(t *testing.T) string {
	t.Helper()
	resetReportManagers(t)
	path := filepath.Join(t.TempDir(), "reports-ext2.mia")
	if err := disk.MakeDisk(disk.MakeDiskOptions{Size: 20, Unit: "M", Path: path}); err != nil {
		t.Fatalf("MakeDisk failed: %v", err)
	}
	if err := partition.Create(partition.CreateOptions{Size: 15, Unit: "M", Path: path, Type: "P", Name: "Part1"}); err != nil {
		t.Fatalf("partition failed: %v", err)
	}
	if _, err := mount.Global.Mount(path, "Part1"); err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	var out bytes.Buffer
	if err := fs.Format(fs.FormatOptions{ID: "961A", Type: "full"}, &out); err != nil {
		t.Fatalf("mkfs failed: %v", err)
	}
	if _, err := session.Login("root", "123", "961A"); err != nil {
		t.Fatalf("login failed: %v", err)
	}
	actor := fs.Actor{User: "root", UID: 1, GID: 1}
	if err := fs.Mkdir(path, currentReportPartitionStart(t), "/home/docs", true, actor); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := fs.Mkfile(path, currentReportPartitionStart(t), "/home/docs/a.txt", false, 75, "", actor); err != nil {
		t.Fatalf("mkfile a failed: %v", err)
	}
	if err := fs.Mkfile(path, currentReportPartitionStart(t), "/home/docs/b.txt", false, 20, "", actor); err != nil {
		t.Fatalf("mkfile b failed: %v", err)
	}
	if err := fs.Mkfile(path, currentReportPartitionStart(t), "/home/docs/large.txt", false, 900, "", actor); err != nil {
		t.Fatalf("mkfile large failed: %v", err)
	}
	_ = session.Logout()
	return path
}

func currentReportPartitionStart(t *testing.T) int64 {
	t.Helper()
	mounted, ok := mount.Global.GetMounted("961A")
	if !ok {
		t.Fatal("missing mount")
	}
	return int64(mounted.Start)
}

func resetReportManagers(t *testing.T) {
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

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
