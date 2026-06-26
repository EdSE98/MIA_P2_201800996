package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"mia_p1_201800996/internal/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	fmt.Printf("MIA Proyecto 2 API escuchando en http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, api.NewRouter()); err != nil {
		log.Fatal(err)
	}
}
