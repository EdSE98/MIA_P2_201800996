package handlers

import (
	"encoding/json"
	"net/http"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/services"
)

func EditFile(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPatch) {
		return
	}
	var req dto.EditFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error("json invalido"))
		return
	}
	if err := services.EditFile(req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, dto.Success("archivo editado", nil))
}

func RenameEntry(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPatch) {
		return
	}
	var req dto.RenameEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error("json invalido"))
		return
	}
	if err := services.RenameEntry(req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, dto.Success("archivo o carpeta renombrado", nil))
}
