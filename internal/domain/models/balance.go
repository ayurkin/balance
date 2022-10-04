package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type Balance struct {
	UserId int64           `json:"user_id"`
	Value  decimal.Decimal `json:"value"`
}

type BalanceWithDesc struct {
	UserId      int64           `json:"user_id"`
	Value       decimal.Decimal `json:"value"`
	Time        time.Time
	Description string `json:"description"`
}
