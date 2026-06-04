package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type User struct {
	UserID   int64
	Login    string
	Password string
}

type UserBalance struct {
	UserID    int64
	Balance   decimal.Decimal
	Withdrawn decimal.Decimal
}

type Order struct {
	OrderID    int64
	UserID     int64
	Number     string
	Status     string
	Accrual    decimal.Decimal
	UploadedAt time.Time
}
