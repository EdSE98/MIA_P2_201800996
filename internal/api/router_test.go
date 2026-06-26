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
