package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/partition"
	"mia_p1_201800996/internal/session"
)

func TestDisksEndpointReturnsJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MIA_DISKS_DIR", dir)

	if err := os.WriteFile(filepath.Join(dir, "api.dsk"), []byte("disk"), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/disks", nil)
	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var response dto.Response
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if !response.OK || response.Message != "discos obtenidos" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestReportFilesServesSVGWithContentType(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MIA_REPORTS_DIR", dir)
	content := `<svg xmlns="http://www.w3.org/2000/svg"></svg>`
	if err := os.WriteFile(filepath.Join(dir, "tree_961A.svg"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/report-files/tree_961A.svg", nil)
	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "image/svg+xml" {
		t.Fatalf("expected SVG content type, got %q", got)
	}
	if rec.Body.String() != content {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestReportFilesRejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MIA_REPORTS_DIR", dir)
	if err := os.WriteFile(filepath.Join(filepath.Dir(dir), "secret.svg"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/report-files/%2e%2e%2fsecret.svg", nil)
	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		t.Fatalf("path traversal unexpectedly succeeded: %s", rec.Body.String())
	}
}

func TestPartitionResizeAndDeleteEndpoints(t *testing.T) {
	path := filepath.Join(t.TempDir(), "api-fdisk.dsk")
	if err := disk.MakeDisk(disk.MakeDiskOptions{Size: 3, Unit: "M", Path: path}); err != nil {
		t.Fatal(err)
	}
	if err := partition.Create(partition.CreateOptions{
		Size: 512, Unit: "K", Path: path, Type: "P", Name: "Part1",
	}); err != nil {
		t.Fatal(err)
	}

	resizeBody, err := json.Marshal(dto.ResizePartitionRequest{
		Path: path, Name: "Part1", Add: 128, Unit: "K",
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPatch, "/api/partitions/resize", bytes.NewReader(resizeBody))
	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("resize expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	deleteBody, err := json.Marshal(dto.DeletePartitionRequest{
		Path: path, Name: "Part1", Delete: "fast",
	})
	if err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest(http.MethodDelete, "/api/partitions", bytes.NewReader(deleteBody))
	rec = httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if _, _, err := partition.SearchPartition(path, "Part1"); err == nil {
		t.Fatal("expected partition deleted through API")
	}
}

func TestFSOperationEndpointsRequireSession(t *testing.T) {
	oldSession := session.Global
	session.Global = session.NewManager()
	t.Cleanup(func() { session.Global = oldSession })

	tests := []struct {
		path string
		body string
	}{
		{"/api/fs/edit", `{"path":"/a.txt","contenido":"/tmp/a.txt"}`},
		{"/api/fs/rename", `{"path":"/a.txt","name":"b.txt"}`},
		{"/api/fs/remove", `{"path":"/a.txt"}`},
	}
	for _, test := range tests {
		method := http.MethodPatch
		if test.path == "/api/fs/remove" {
			method = http.MethodDelete
		}
		req := httptest.NewRequest(method, test.path, bytes.NewBufferString(test.body))
		rec := httptest.NewRecorder()
		NewRouter().ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s expected 400, got %d: %s", test.path, rec.Code, rec.Body.String())
		}
		var response dto.Response
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}
		if response.OK || response.Error != "necesita iniciar sesion" {
			t.Fatalf("%s unexpected response: %+v", test.path, response)
		}
	}
}
