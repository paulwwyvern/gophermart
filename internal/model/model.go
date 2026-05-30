package model

import "time"

type User struct {
	UserID   int64
	Login    string
	Password string
}

type Order struct {
	OrderID    int64
	UserID     int64
	Number     string
	Status     string
	Accrual    int
	UploadedAt time.Time
}
