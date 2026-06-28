package handlers

import (
	"encoding/json"
	"net/http"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/services"
)

func ExecuteCommand(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}
	var req dto.ExecuteCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error("json invalido"))
		return
	}
	result, err := services.ExecuteCommand(req.Command)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, dto.Success("comando ejecutado", result))
}
