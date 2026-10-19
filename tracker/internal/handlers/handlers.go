package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"personal-expenses/internal/models"
	"personal-expenses/internal/repository"
	"strconv"
	"time"
)

var repo *repository.ExpenseRepository

func SetRepository(r *repository.ExpenseRepository) {
	repo = r
}

var tmpl = template.Must(template.ParseFiles(
	filepath.Join("web", "templates", "layout.html"),
	filepath.Join("web", "templates", "index.html"),
))

func PingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("pong"))
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Веб-приложение 'Учёт личных трат' — семестровый проект на Go."))
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	list, err := repo.GetAll()
	if err != nil {
		http.Error(w, "Ошибка получения данных из БД", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":    "Главная",
		"Expenses": list,
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
		return
	}
}

func AddExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	amountStr := r.FormValue("amount")
	description := r.FormValue("description")
	dateStr := r.FormValue("date")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Некорректная сумма", http.StatusBadRequest)
		return
	}

	if description == "" {
		http.Error(w, "Описание не может быть пустым", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Неверный формат даты", http.StatusBadRequest)
		return
	}

	err = repo.Create(models.Expense{
		Amount:      amount,
		Description: description,
		Date:        date,
	})
	if err != nil {
		http.Error(w, "Ошибка сохранения в БД", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
