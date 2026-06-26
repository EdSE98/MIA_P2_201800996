package handlers

import (
	"encoding/json"
	"net/http"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/services"
)

func Disks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		disks, err := services.ListDisks()
		if err != nil {
			writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
			return
		}
		writeJSON(w, http.StatusOK, dto.Success("discos obtenidos", disks))
	case http.MethodPost:
		var req dto.CreateDiskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, dto.Error("json invalido"))
			return
		}
		disk, err := services.CreateDisk(req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
			return
		}
		writeJSON(w, http.StatusOK, dto.Success("disco creado", disk))
	case http.MethodDelete:
		var req dto.DeleteDiskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, dto.Error("json invalido"))
			return
		}
		if err := services.DeleteDisk(req.Path); err != nil {
			writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
			return
		}
		writeJSON(w, http.StatusOK, dto.Success("disco eliminado", nil))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, dto.Error("metodo no permitido"))
	}
}
