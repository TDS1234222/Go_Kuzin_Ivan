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
	db, err := sql.Open("sqlite", "expenses.db")
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		log.Fatalf("Ошибка настройки WAL: %v", err)
	}

	initSQL, err := os.ReadFile(filepath.Join("migrations", "init.sql"))
	if err != nil {
		log.Fatalf("Ошибка чтения файла миграций: %v", err)
	}

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

	// Наш Middleware для проверки авторизации (Пункт 3)
	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil {
				http.Redirect(w, r, "/auth", http.StatusSeeOther)
				return
			}
			_, err = repo.GetUserIDBySession(cookie.Value)
			if err != nil {
				http.Redirect(w, r, "/auth", http.StatusSeeOther)
				return
			}
			next(w, r)
		}
	}

	mux := http.NewServeMux()
	
	// Защищенные роуты (оборачиваем в middleware)
	mux.HandleFunc("/", authMiddleware(h.Index))
	mux.HandleFunc("/add", authMiddleware(h.AddExpense))
	mux.HandleFunc("/delete", authMiddleware(h.DeleteExpense))
	mux.HandleFunc("/edit", authMiddleware(h.EditPage))
	mux.HandleFunc("/update", authMiddleware(h.UpdateExpense))
	
	// Открытые роуты
	mux.HandleFunc("/auth", h.AuthPage)
	mux.HandleFunc("/register", h.Register)
	mux.HandleFunc("/login", h.Login)
	mux.HandleFunc("/logout", h.Logout)
	mux.HandleFunc("/about", h.About)
	mux.HandleFunc("/ping", h.Ping)

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
