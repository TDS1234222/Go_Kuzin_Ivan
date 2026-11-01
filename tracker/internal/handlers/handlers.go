package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"net/http"
	"path/filepath"
	"personal-expenses/internal/models"
	"personal-expenses/internal/repository"
	"strconv"
	"time"
)

type Handler struct {
	repo *repository.ExpenseRepository
}

func New(repo *repository.ExpenseRepository) *Handler {
	return &Handler{repo: repo}
}

var indexTmpl = template.Must(template.ParseFiles(
	filepath.Join("web", "templates", "layout.html"),
	filepath.Join("web", "templates", "index.html"),
))

var editTmpl = template.Must(template.ParseFiles(
	filepath.Join("web", "templates", "layout.html"),
	filepath.Join("web", "templates", "edit.html"),
))

var authTmpl = template.Must(template.ParseFiles(
	filepath.Join("web", "templates", "layout.html"),
	filepath.Join("web", "templates", "auth.html"),
))

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func (h *Handler) GetUserID(r *http.Request) (int, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return 0, err
	}
	return h.repo.GetUserIDBySession(cookie.Value)
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("pong"))
}

func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Веб-приложение 'Учёт личных трат' — семестровый проект на Go."))
}

func (h *Handler) AuthPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	data := map[string]interface{}{"Title": "Вход / Регистрация"}
	authTmpl.ExecuteTemplate(w, "layout", data)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		http.Error(w, "Заполните все поля", http.StatusBadRequest)
		return
	}

	err := h.repo.CreateUser(email, hashPassword(password))
	if err != nil {
		http.Error(w, "Пользователь с таким email уже существует", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/auth", http.StatusSeeOther)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.repo.GetUserByEmail(email)
	if err != nil || user.PasswordHash != hashPassword(password) {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	b := make([]byte, 16)
	rand.Read(b)
	sessionID := hex.EncodeToString(b)

	err = h.repo.CreateSession(sessionID, user.ID)
	if err != nil {
		http.Error(w, "Ошибка создания сессии", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		h.repo.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/auth", http.StatusSeeOther)
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := h.GetUserID(r)

	// Получаем query-параметры фильтрации (Пункт 4 ТЗ)
	catIDStr := r.URL.Query().Get("category_id")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	var catID int
	if catIDStr != "" {
		catID, _ = strconv.Atoi(catIDStr)
	}

	// Запрашиваем отфильтрованный или полный список трат (Пункт 5 ТЗ)
	list, err := h.repo.GetAll(userID, catID, dateFrom, dateTo)
	if err != nil {
		http.Error(w, "Ошибка получения данных из БД", http.StatusInternalServerError)
		return
	}

	// Запрашиваем категории для селектов
	cats, err := h.repo.GetCategories()
	if err != nil {
		http.Error(w, "Ошибка получения категорий", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":             "Главная",
		"Expenses":          list,
		"Categories":        cats,
		"CurrentCategoryID": catID,
		"CurrentDateFrom":   dateFrom,
		"CurrentDateTo":     dateTo,
		"IsAuth":            true,
	}

	indexTmpl.ExecuteTemplate(w, "layout", data)
}

func (h *Handler) AddExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := h.GetUserID(r)

	r.ParseForm()
	amountStr := r.FormValue("amount")
	description := r.FormValue("description")
	categoryIDStr := r.FormValue("category_id")
	dateStr := r.FormValue("date")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Некорректная сумма", http.StatusBadRequest)
		return
	}

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID <= 0 {
		http.Error(w, "Некорректная категория", http.StatusBadRequest)
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

	err = h.repo.Create(models.Expense{
		UserID:      userID,
		CategoryID:  categoryID,
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

func (h *Handler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := h.GetUserID(r)

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	err = h.repo.Delete(id, userID)
	if err != nil {
		http.Error(w, "Ошибка удаления из БД", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) EditPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := h.GetUserID(r)

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	exp, err := h.repo.GetByID(id, userID)
	if err != nil {
		http.Error(w, "Расход не найден", http.StatusNotFound)
		return
	}

	cats, err := h.repo.GetCategories()
	if err != nil {
		http.Error(w, "Ошибка получения категорий", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":      "Редактирование",
		"Expense":    exp,
		"Categories": cats,
		"IsAuth":     true,
	}

	editTmpl.ExecuteTemplate(w, "layout", data)
}

func (h *Handler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := h.GetUserID(r)

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	amountStr := r.FormValue("amount")
	description := r.FormValue("description")
	categoryIDStr := r.FormValue("category_id")
	dateStr := r.FormValue("date")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Некорректная сумма", http.StatusBadRequest)
		return
	}

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID <= 0 {
		http.Error(w, "Некорректная категория", http.StatusBadRequest)
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

	err = h.repo.Update(models.Expense{
		ID:          id,
		UserID:      userID,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Date:        date,
	})
	if err != nil {
		http.Error(w, "Ошибка обновления в БД", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
