package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"mia_p1_201800996/internal/api/dto"
	"mia_p1_201800996/internal/api/handlers"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handlers.Health)
	mux.HandleFunc("/api/mounts", handlers.Mounts)
	mux.HandleFunc("/api/login", handlers.Login)
	mux.HandleFunc("/api/logout", handlers.Logout)
	mux.HandleFunc("/api/disks", handlers.Disks)
	mux.HandleFunc("/api/partitions", handlers.Partitions)
	mux.HandleFunc("/api/mount", handlers.Mount)
	mux.HandleFunc("/api/unmount", handlers.Unmount)
	mux.HandleFunc("/api/mkfs", handlers.Mkfs)
	mux.HandleFunc("/api/reports", handlers.Reports)

	return recoverJSON(cors(mux))
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func recoverJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(dto.Error(fmt.Sprintf("error interno: %v", recovered)))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
