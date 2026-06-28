package handlers

import (
	"encoding/json"
	"net/http"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/services"
)

func Partitions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		partitions, err := services.ListPartitions(r.URL.Query().Get("path"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
			return
		}
		writeJSON(w, http.StatusOK, dto.Success("particiones obtenidas", partitions))
	case http.MethodPost:
		var req dto.CreatePartitionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, dto.Error("json invalido"))
			return
		}
		if err := services.CreatePartition(req); err != nil {
			writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
			return
		}
		writeJSON(w, http.StatusOK, dto.Success("particion creada", nil))
	case http.MethodDelete:
		var req dto.DeletePartitionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, dto.Error("json invalido"))
			return
		}
		if err := services.DeletePartition(req); err != nil {
			writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
			return
		}
		writeJSON(w, http.StatusOK, dto.Success("particion eliminada", nil))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, dto.Error("metodo no permitido"))
	}
}

func ResizePartition(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPatch) {
		return
	}
	var req dto.ResizePartitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error("json invalido"))
		return
	}
	if err := services.ResizePartition(req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, dto.Success("particion redimensionada", nil))
}
