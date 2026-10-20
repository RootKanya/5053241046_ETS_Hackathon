package main

import (
	"log"
	"net/http"
)

func main() {
	handler := NewHandler()
	log.Println("Server PinjamKampus Persona B berjalan di http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Server gagal berjalan: %v", err)
	}
}