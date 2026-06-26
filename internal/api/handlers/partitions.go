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
	default:
		writeJSON(w, http.StatusMethodNotAllowed, dto.Error("metodo no permitido"))
	}
}
