package ports

import (
	"balance/internal/domain/models"
	"context"
)

type BalancePort interface {
	AddIncome(ctx context.Context, income models.BalanceWithDesc) error
	AddExpense(ctx context.Context, expense models.BalanceWithDesc) error
	DoTransfer(ctx context.Context, transaction models.Transaction) error
	GetBalance(ctx context.Context, userId int64) (models.Balance, error)
}
