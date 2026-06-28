package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/services"
)

func Reports(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}
	var req dto.ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error("json invalido"))
		return
	}
	report, err := services.GenerateReport(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, dto.Success("reporte generado", report))
}

func ReportFiles(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}

	const prefix = "/api/report-files/"
	filename, err := url.PathUnescape(strings.TrimPrefix(r.URL.EscapedPath(), prefix))
	if err != nil || filename == "" {
		writeJSON(w, http.StatusBadRequest, dto.Error("nombre de archivo de reporte invalido"))
		return
	}

	reportPath, contentType, err := services.ResolveReportFile(filename)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, services.ErrReportFileNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, dto.Error(err.Error()))
		return
	}

	file, err := os.Open(reportPath)
	if err != nil {
		writeJSON(w, http.StatusNotFound, dto.Error("archivo de reporte no encontrado"))
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.Error("no se pudo leer el reporte"))
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", `inline; filename="`+filepath.Base(reportPath)+`"`)
	http.ServeContent(w, r, info.Name(), info.ModTime().Truncate(time.Second), file)
}
