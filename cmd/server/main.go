package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"mia_p1_201800996/internal/api"
)

func main() {
	addr := strings.TrimSpace(os.Getenv("MIA_API_ADDR"))
	if addr == "" {
		port := strings.TrimSpace(os.Getenv("PORT"))
		if port == "" {
			port = "8080"
		}
		addr = "127.0.0.1:" + port
	}

	fmt.Printf("MIA Proyecto 2 API escuchando en http://%s\n", addr)
	if err := http.ListenAndServe(addr, api.NewRouter()); err != nil {
		log.Fatal(err)
	}
}
