package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type Transaction struct {
	Id          int64           `json:"id"`
	UserIdFrom  int64           `json:"user_id_from"`
	UserIdTo    int64           `json:"user_id_to"`
	Value       decimal.Decimal `json:"value"`
	Time        time.Time
	Description string `json:"description"`
}
