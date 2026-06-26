package handlers

import (
	"net/http"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/services"
)

func FSList(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	data, err := services.ListFS(r.URL.Query().Get("id"), r.URL.Query().Get("path"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, dto.Success("directorio obtenido", data))
}

func FSRead(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	data, err := services.ReadFS(r.URL.Query().Get("id"), r.URL.Query().Get("path"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, dto.Success("archivo leído", data))
}

func FSStat(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	data, err := services.StatFS(r.URL.Query().Get("id"), r.URL.Query().Get("path"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.Error(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, dto.Success("metadata obtenida", data))
}
