package services

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/reports"
)

const defaultReportsDir = "/home/eduardo/parte2/reportes/api"

var unsafeReportName = regexp.MustCompile(`[^A-Za-z0-9_-]+`)
var ErrReportFileNotFound = errors.New("archivo de reporte no encontrado")

func ReportsDir() string {
	if value := strings.TrimSpace(os.Getenv("MIA_REPORTS_DIR")); value != "" {
		return value
	}
	return defaultReportsDir
}

func GenerateReport(req dto.ReportRequest) (dto.ReportResponse, error) {
	if strings.TrimSpace(req.ID) == "" {
		return dto.ReportResponse{}, fmt.Errorf("id es obligatorio")
	}
	if strings.TrimSpace(req.Name) == "" {
		return dto.ReportResponse{}, fmt.Errorf("name es obligatorio")
	}

	reportName := strings.ToLower(strings.TrimSpace(req.Name))
	if reportName == "bm_bloc" {
		reportName = "bm_block"
	}
	format := normalizeReportFormat(reportName, req.Format)
	outputPath := filepath.Join(ReportsDir(), reportFileName(reportName, req.ID, format))

	params := map[string]string{
		"id":   req.ID,
		"name": reportName,
		"path": outputPath,
	}
	if strings.TrimSpace(req.PathFileLs) != "" {
		params["path_file_ls"] = req.PathFileLs
	}

	var out bytes.Buffer
	if err := reports.Generate(params, &out); err != nil {
		return dto.ReportResponse{}, err
	}

	finalPath := outputPath
	if _, err := os.Stat(finalPath); err != nil {
		dotPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".dot"
		if _, dotErr := os.Stat(dotPath); dotErr == nil {
			finalPath = dotPath
		}
	}

	return dto.ReportResponse{
		Name:        reportName,
		Path:        finalPath,
		URL:         "/api/report-files/" + url.PathEscape(filepath.Base(finalPath)),
		ContentType: contentTypeForPath(finalPath),
	}, nil
}

func ResolveReportFile(filename string) (string, string, error) {
	filename = strings.TrimSpace(filename)
	if filename == "" ||
		filepath.IsAbs(filename) ||
		filename != filepath.Base(filename) ||
		strings.ContainsAny(filename, `/\`+"\x00") {
		return "", "", fmt.Errorf("nombre de archivo de reporte invalido")
	}

	baseDir, err := filepath.Abs(ReportsDir())
	if err != nil {
		return "", "", fmt.Errorf("resolver carpeta de reportes: %w", err)
	}
	targetPath := filepath.Join(baseDir, filename)
	relative, err := filepath.Rel(baseDir, targetPath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", "", fmt.Errorf("archivo fuera de la carpeta de reportes")
	}

	info, err := os.Lstat(targetPath)
	if errors.Is(err, os.ErrNotExist) {
		return "", "", ErrReportFileNotFound
	}
	if err != nil {
		return "", "", fmt.Errorf("leer archivo de reporte: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return "", "", fmt.Errorf("archivo de reporte invalido")
	}

	return targetPath, contentTypeForPath(targetPath), nil
}

func normalizeReportFormat(reportName string, format string) string {
	normalized := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(format), "."))
	if normalized == "" {
		if reportName == "file" {
			return "txt"
		}
		return "svg"
	}
	if normalized == "jpeg" {
		return "jpg"
	}
	return normalized
}

func reportFileName(name string, id string, format string) string {
	cleanName := unsafeReportName.ReplaceAllString(name, "_")
	cleanID := unsafeReportName.ReplaceAllString(id, "_")
	return fmt.Sprintf("%s_%s.%s", cleanName, cleanID, format)
}

func contentTypeForPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}
