package models

import "time"

type Expense struct {
	Amount      float64
	Description string
	Date        time.Time
}
