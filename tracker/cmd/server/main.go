package main

import (
	"log"
	"net/http"
	"personal-expenses/internal/handlers"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.IndexHandler)
	mux.HandleFunc("/about", handlers.AboutHandler)
	mux.HandleFunc("/ping", handlers.PingHandler)

	log.Println("Успешно: Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
