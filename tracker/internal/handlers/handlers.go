package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"personal-expenses/internal/models"
	"strconv"
	"sync"
	"time"
)

var (
	Expenses []models.Expense
	mu       sync.Mutex
)

func init() {
	Expenses = []models.Expense{
		{Amount: 450.00, Description: "Обед в столовой", Date: time.Now()},
		{Amount: 120.00, Description: "Проезд на метро", Date: time.Now()},
	}
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 Method Not Allowed"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Веб-приложение 'Учёт личных трат' — семестровый проект на Go."))
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		renderIndex(w)
	case http.MethodPost:
		addExpense(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func renderIndex(w http.ResponseWriter) {
	layoutPath := filepath.Join("web", "templates", "layout.html")
	indexPath := filepath.Join("web", "templates", "index.html")

	tmpl, err := template.ParseFiles(layoutPath, indexPath)
	if err != nil {
		http.Error(w, "Ошибка чтения шаблонов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	mu.Lock()
	data := map[string]interface{}{
		"Title":    "Главная страница",
		"Expenses": Expenses,
	}
	mu.Unlock()

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, "Ошибка отрисовки: "+err.Error(), http.StatusInternalServerError)
	}
}

func addExpense(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}

	amountStr := r.FormValue("amount")
	description := r.FormValue("description")
	dateStr := r.FormValue("date")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Сумма должна быть больше нуля", http.StatusBadRequest)
		return
	}

	if description == "" {
		http.Error(w, "Описание пустое", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	mu.Lock()
	Expenses = append(Expenses, models.Expense{
		Amount:      amount,
		Description: description,
		Date:        date,
	})
	mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
