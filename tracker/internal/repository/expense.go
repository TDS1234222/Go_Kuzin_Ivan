package repository

import (
	"database/sql"
	"personal-expenses/internal/models"
	"time"
)

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) CreateUser(email, passwordHash string) error {
	_, err := r.db.Exec(`INSERT INTO users (email, password_hash) VALUES (?, ?)`, email, passwordHash)
	return err
}

func (r *ExpenseRepository) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	err := r.db.QueryRow(`SELECT id, email, password_hash FROM users WHERE email = ?`, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash)
	return user, err
}

func (r *ExpenseRepository) CreateSession(sessionID string, userID int) error {
	_, err := r.db.Exec(`INSERT INTO sessions (id, user_id) VALUES (?, ?)`, sessionID, userID)
	return err
}

func (r *ExpenseRepository) GetUserIDBySession(sessionID string) (int, error) {
	var userID int
	err := r.db.QueryRow(`SELECT user_id FROM sessions WHERE id = ?`, sessionID).Scan(&userID)
	return userID, err
}

func (r *ExpenseRepository) DeleteSession(sessionID string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id = ?`, sessionID)
	return err
}

func (r *ExpenseRepository) GetCategories() ([]models.Category, error) {
	query := `SELECT id, name FROM categories ORDER BY id ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func (r *ExpenseRepository) Create(exp models.Expense) error {
	query := `INSERT INTO expenses (user_id, category_id, amount, description, date) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, exp.UserID, exp.CategoryID, exp.Amount, exp.Description, exp.Date.Format("2006-01-02"))
	return err
}

// Вывод списка трат через JOIN с фильтрацией по категории и диапазону дат (Пункт 3, 4, 5)
func (r *ExpenseRepository) GetAll(userID int, catID int, dateFrom, dateTo string) ([]models.Expense, error) {
	var rows *sql.Rows
	var err error

	// Базовый SQL-запрос с JOIN (Пункт 3 ТЗ)
	query := `
		SELECT e.id, e.user_id, e.category_id, c.name, e.amount, e.description, e.date 
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id
		WHERE e.user_id = ?
	`
	var args []interface{}
	args = append(args, userID)

	// Пункт 4 ТЗ: Динамически дописываем фильтры, если они переданы
	if catID > 0 {
		query += ` AND e.category_id = ?`
		args = append(args, catID)
	}
	if dateFrom != "" {
		query += ` AND e.date >= ?`
		args = append(args, dateFrom)
	}
	if dateTo != "" {
		query += ` AND e.date <= ?`
		args = append(args, dateTo)
	}

	query += ` ORDER BY e.date DESC`

	rows, err = r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Expense
	for rows.Next() {
		var exp models.Expense
		var dateStr string

		err := rows.Scan(&exp.ID, &exp.UserID, &exp.CategoryID, &exp.CategoryName, &exp.Amount, &exp.Description, &dateStr)
		if err != nil {
			return nil, err
		}

		exp.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, err
		}

		list = append(list, exp)
	}

	return list, nil
}

func (r *ExpenseRepository) Delete(id, userID int) error {
	_, err := r.db.Exec(`DELETE FROM expenses WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

func (r *ExpenseRepository) GetByID(id, userID int) (models.Expense, error) {
	var exp models.Expense
	var dateStr string
	query := `
		SELECT e.id, e.user_id, e.category_id, c.name, e.amount, e.description, e.date 
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id
		WHERE e.id = ? AND e.user_id = ?
	`
	err := r.db.QueryRow(query, id, userID).Scan(&exp.ID, &exp.UserID, &exp.CategoryID, &exp.CategoryName, &exp.Amount, &exp.Description, &dateStr)
	if err != nil {
		return exp, err
	}
	exp.Date, err = time.Parse("2006-01-02", dateStr)
	return exp, err
}

func (r *ExpenseRepository) Update(exp models.Expense) error {
	query := `UPDATE expenses SET amount = ?, description = ?, category_id = ?, date = ? WHERE id = ? AND user_id = ?`
	_, err := r.db.Exec(query, exp.Amount, exp.Description, exp.CategoryID, exp.Date.Format("2006-01-02"), exp.ID, exp.UserID)
	return err
}
