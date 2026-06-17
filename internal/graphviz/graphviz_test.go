package graphviz

import "testing"

func TestFormatFromExtension(t *testing.T) {
	tests := map[string]string{
		"/tmp/a.jpg":  "jpg",
		"/tmp/a.jpeg": "jpg",
		"/tmp/a.png":  "png",
		"/tmp/a.pdf":  "pdf",
		"/tmp/a.svg":  "svg",
		"/tmp/a.dot":  "dot",
	}

	for path, want := range tests {
		t.Run(path, func(t *testing.T) {
			got, err := FormatFromExtension(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != want {
				t.Fatalf("format = %q, want %q", got, want)
			}
		})
	}
}

func TestFormatFromExtensionRejectsInvalid(t *testing.T) {
	if _, err := FormatFromExtension("/tmp/a.bmp"); err == nil {
		t.Fatal("expected invalid extension error")
	}
}

func TestIsDotAvailableDoesNotPanic(t *testing.T) {
	_ = IsDotAvailable()
}
