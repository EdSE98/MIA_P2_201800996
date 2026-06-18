package reports

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"mia_p1_201800996/internal/graphviz"
)

func writeDotAndRender(outputPath string, dot string) (string, error) {
	if _, err := graphviz.FormatFromExtension(outputPath); err != nil {
		return "", err
	}

	dotPath := outputPath
	if strings.ToLower(filepath.Ext(outputPath)) != ".dot" {
		dotPath = strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".dot"
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(dotPath, []byte(dot), 0o644); err != nil {
		return "", err
	}

	if strings.ToLower(filepath.Ext(outputPath)) == ".dot" {
		return dotPath, nil
	}
	if err := graphviz.RenderDot(dotPath, outputPath); err != nil {
		return dotPath, fmt.Errorf("%w; se conservo el archivo DOT en %s", err, dotPath)
	}
	return dotPath, nil
}

func esc(value string) string {
	return html.EscapeString(value)
}

func byteText(value byte) string {
	if value == 0 {
		return ""
	}
	return string([]byte{value})
}

func pct(size int64, total int64) string {
	if total <= 0 {
		return "0.00%"
	}
	return fmt.Sprintf("%.2f%%", float64(size)*100/float64(total))
}

func htmlCell(parts ...string) string {
	return "<TD>" + strings.Join(parts, "<BR/>") + "</TD>"
}

func htmlLabel(value string) string {
	escaped := esc(value)
	escaped = strings.ReplaceAll(escaped, "&lt;br/&gt;", "<BR/>")
	escaped = strings.ReplaceAll(escaped, "&lt;BR/&gt;", "<BR/>")
	return escaped
}
