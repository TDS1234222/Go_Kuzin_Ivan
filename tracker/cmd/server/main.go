package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"personal-expenses/internal/handlers"
	"personal-expenses/internal/repository"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "expenses.db")
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	initSQL, err := os.ReadFile("migrations/init.sql")
	if err != nil {
		log.Fatalf("Ошибка чтения файла миграций: %v", err)
	}

	_, err = db.Exec(string(initSQL))
	if err != nil {
		log.Fatalf("Ошибка инициализации таблицы: %v", err)
	}

	repo := repository.NewExpenseRepository(db)
	handlers.SetRepository(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.IndexHandler)
	mux.HandleFunc("/add", handlers.AddExpenseHandler)
	mux.HandleFunc("/about", handlers.AboutHandler)
	mux.HandleFunc("/ping", handlers.PingHandler)

	log.Println("Сервер запущен на http://localhost:8080")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
