package models

import (
	"github.com/shopspring/decimal"
)

type Balance struct {
	UserId  int64           `json:"user_id"`
	Balance decimal.Decimal `json:"balance"`
}
