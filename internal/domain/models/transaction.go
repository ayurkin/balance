package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type Transaction struct {
	Id          int64           `json:"id"`
	FromId      int64           `json:"from_id"`
	ToId        int64           `json:"to_id"`
	Amount      decimal.Decimal `json:"amount"`
	Time        time.Time       `json:"time"`
	Description string          `json:"description"`
}
