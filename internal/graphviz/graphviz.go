package graphviz

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func IsDotAvailable() bool {
	_, err := exec.LookPath("dot")
	return err == nil
}

func RenderDot(dotPath string, outputPath string) error {
	format, err := FormatFromExtension(outputPath)
	if err != nil {
		return err
	}
	if format == "dot" {
		return nil
	}
	if !IsDotAvailable() {
		return fmt.Errorf("Graphviz no esta instalado o 'dot' no esta en PATH")
	}

	cmd := exec.Command("dot", "-T"+format, dotPath, "-o", outputPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error ejecutando dot: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func FormatFromExtension(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "jpg", nil
	case ".png":
		return "png", nil
	case ".pdf":
		return "pdf", nil
	case ".svg":
		return "svg", nil
	case ".dot":
		return "dot", nil
	default:
		return "", fmt.Errorf("extension de reporte no soportada %q", ext)
	}
}
