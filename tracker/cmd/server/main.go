package main

import (
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"personal-expenses/internal/handlers"
	"personal-expenses/internal/repository"

	_ "modernc.org/sqlite"
)

func main() {
	// Открываем базу данных в корне проекта
	db, err := sql.Open("sqlite", "expenses.db")
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	// Настраиваем пул и WAL-режим для потокобезопасности SQLite
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		log.Fatalf("Ошибка настройки WAL: %v", err)
	}

	// Читаем файл миграций из папки migrations
	initSQL, err := os.ReadFile(filepath.Join("migrations", "init.sql"))
	if err != nil {
		log.Fatalf("Ошибка чтения файла миграций: %v", err)
	}

	// Выполняем SQL-команды построчно, так как драйвер не умеет в мультизапросы
	statements := strings.Split(string(initSQL), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err = db.Exec(stmt); err != nil {
			log.Fatalf("Ошибка выполнения миграции [%s]: %v", stmt, err)
		}
	}

	repo := repository.NewExpenseRepository(db)
	h := handlers.New(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.Index)
	mux.HandleFunc("/add", h.AddExpense)
	mux.HandleFunc("/about", h.About)
	mux.HandleFunc("/ping", h.Ping)

	// Занимаем порт
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Не удалось занять порт 8080: %v", err)
	}

	log.Println("Сервер успешно запущен на http://localhost:8080")
	
	err = http.Serve(ln, mux)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Критическая ошибка сервера: %v", err)
	}
}
