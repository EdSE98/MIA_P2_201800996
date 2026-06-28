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

func RemoveEntry(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodDelete) {
		return
	}
	var req dto.RemoveEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error("json invalido"))
		return
	}
	if err := services.RemoveEntry(req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, dto.Success("archivo o carpeta eliminado", nil))
}

func CopyEntry(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}
	var req dto.TransferEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error("json invalido"))
		return
	}
	result, err := services.CopyEntry(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, dto.Success("archivo o carpeta copiado", result))
}

func MoveEntry(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPatch) {
		return
	}
	var req dto.TransferEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error("json invalido"))
		return
	}
	if err := services.MoveEntry(req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, dto.Success("archivo o carpeta movido", nil))
}
