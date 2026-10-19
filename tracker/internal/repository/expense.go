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

func (r *ExpenseRepository) Create(exp models.Expense) error {
	query := `INSERT INTO expenses (amount, description, date) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, exp.Amount, exp.Description, exp.Date.Format("2006-01-02"))
	return err
}

func (r *ExpenseRepository) GetAll() ([]models.Expense, error) {
	query := `SELECT amount, description, date FROM expenses ORDER BY date DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Expense
	for rows.Next() {
		var exp models.Expense
		var dateStr string
		
		err := rows.Scan(&exp.Amount, &exp.Description, &dateStr)
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
