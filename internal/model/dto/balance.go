package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type GetUserBalanceResponse struct {
	Balance   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

type CreateWithdrawalRequest struct {
	UserID      int64           `json:"-"`
	OrderNumber string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
}

type GetWithdrawalsResponse struct {
	OrderNumber string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at"`
}
