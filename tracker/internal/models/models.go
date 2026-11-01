package models

import "time"

type User struct {
	ID           int
	Email        string
	PasswordHash string
}

type Category struct {
	ID   int
	Name string
}

type Expense struct {
	ID           int
	UserID       int
	CategoryID   int
	CategoryName string
	Amount       float64
	Description  string
	Date         time.Time
}
