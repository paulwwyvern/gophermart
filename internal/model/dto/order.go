package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type GetOrdersResponse struct {
	Number     string          `json:"number"`
	Status     string          `json:"status"`
	Accrual    decimal.Decimal `json:"accrual,omitzero"`
	UploadedAt time.Time       `json:"uploaded_at"`
}
