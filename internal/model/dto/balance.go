package dto

import "github.com/shopspring/decimal"

type GetUserBalanceResponse struct {
	Balance   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}
