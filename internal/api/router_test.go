package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"mia_p1_201800996/internal/api/dto"
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
