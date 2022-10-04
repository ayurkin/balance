package postgres

import (
	"balance/internal/domain/errors"
	"balance/internal/domain/models"
	"context"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

func (db *Database) AddIncome(ctx context.Context, income models.BalanceWithDesc) error {
	tx, err := db.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO balance.history
				(from_id, to_id, value, occurred_at, description)
			VALUES
				($1, $2, $3, $4, $5)`,
		0, income.UserId, income.Value, income.Time, income.Description)

	if err != nil {
		return fmt.Errorf("add transaction to history query exec failed: %w", err)
	}

	var isUserIdExist bool

	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT user_id FROM balance.balance WHERE user_id = $1) AS exists",
		income.UserId).Scan(&isUserIdExist)

	if err != nil {
		return fmt.Errorf("check user_id exists query row failed: %w", err)
	}
	if isUserIdExist {
		_, err = tx.Exec(ctx,
			"UPDATE balance.balance SET value = value + $1 WHERE user_id = $2",
			income.Value, income.UserId)
		if err != nil {
			return fmt.Errorf("add income query exec failed: %w", err)
		}
	} else {
		_, err = tx.Exec(ctx,
			"INSERT INTO balance.balance (user_id, value) VALUES($1, $2)",
			income.UserId, income.Value)
		if err != nil {
			return fmt.Errorf("add new user_id with balance query exec failed: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx commit failed failed: %w", err)
	}
	return nil
}

func (db *Database) AddExpense(ctx context.Context, expense models.BalanceWithDesc) error {
	tx, err := db.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO balance.history
				(from_id, to_id, value, occurred_at, description)
			VALUES
				($1, $2, $3, $4, $5)`,
		expense.UserId, 0, expense.Value, expense.Time, expense.Description)
	if err != nil {
		return fmt.Errorf("add transaction to history query exec failed: %w", err)
	}

	var isUserIdExist bool

	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT user_id FROM balance.balance WHERE user_id = $1) AS exists",
		expense.UserId).Scan(&isUserIdExist)

	if err != nil {
		return fmt.Errorf("check user_id exists query row failed: %w", err)
	}
	if isUserIdExist {
		_, err = tx.Exec(ctx,
			"UPDATE balance.balance SET value = value - $1 WHERE user_id = $2",
			expense.Value, expense.UserId)
		if err != nil {

			if errPq, ok := err.(*pgconn.PgError); ok {
				if errPq.Code == pgerrcode.CheckViolation {
					return fmt.Errorf("user_id %d: %w", expense.UserId, errors.NotEnoughUserBalanceError)
				}
			}

			return fmt.Errorf("add expense query exec failed: %w", err)
		}
	} else {
		return fmt.Errorf("user_id %d: %w", expense.UserId, errors.UnknownUserIdError)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx commit failed failed: %w", err)
	}
	return nil
}

func (db *Database) DoTransfer(ctx context.Context, transaction models.Transaction) error {
	tx, err := db.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO balance.history
				(from_id, to_id, value, occurred_at, description)
			VALUES
				($1, $2, $3, $4, $5)`,
		transaction.UserIdFrom, transaction.UserIdTo, transaction.Value,
		transaction.Time, transaction.Description)
	if err != nil {
		return fmt.Errorf("add transaction to history query exec failed: %w", err)
	}

	var isUserIdFromExist bool

	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT user_id FROM balance.balance WHERE user_id = $1) AS exists",
		transaction.UserIdFrom).Scan(&isUserIdFromExist)

	if err != nil {
		return fmt.Errorf("check user_id exists query row failed: %w", err)
	}
	if isUserIdFromExist {
		_, err = tx.Exec(ctx,
			"UPDATE balance.balance SET value = value - $1 WHERE user_id = $2",
			transaction.Value, transaction.UserIdFrom)
		if err != nil {

			if errPq, ok := err.(*pgconn.PgError); ok {
				if errPq.Code == pgerrcode.CheckViolation {
					return fmt.Errorf("user_id %d: %w", transaction.UserIdFrom, errors.NotEnoughUserBalanceError)
				}
			}

			return fmt.Errorf("add expense query exec failed: %v", err)
		}
	} else {
		return fmt.Errorf("user_id %d: %w", transaction.UserIdFrom, errors.UnknownUserIdError)
	}

	var isUserIdToExist bool

	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT user_id FROM balance.balance WHERE user_id = $1) AS exists",
		transaction.UserIdTo).Scan(&isUserIdToExist)

	if err != nil {
		return fmt.Errorf("check user_id exists query row failed: %w", err)
	}
	if isUserIdToExist {
		_, err = tx.Exec(ctx,
			"UPDATE balance.balance SET value = value + $1 WHERE user_id = $2",
			transaction.Value, transaction.UserIdTo)
		if err != nil {
			return fmt.Errorf("add income query exec failed: %v", err)
		}
	} else {
		_, err = tx.Exec(ctx,
			"INSERT INTO balance.balance (user_id, value) VALUES($1, $2)",
			transaction.UserIdTo, transaction.Value)
		if err != nil {
			return fmt.Errorf("create new user_id with balance query exec failed: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx commit failed failed: %w", err)
	}
	return nil
}

func (db *Database) GetBalance(ctx context.Context, userId int64) (models.Balance, error) {
	var balanceValue string
	var isUserIdExist bool

	err := db.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT user_id FROM balance.balance WHERE user_id = $1) AS exists",
		userId).Scan(&isUserIdExist)
	if err != nil {
		return models.Balance{}, fmt.Errorf("check user_id exists query row failed: %w", err)
	}
	if isUserIdExist {
		err = db.DB.QueryRow(ctx,
			"SELECT value FROM balance.balance WHERE user_id = $1", userId).Scan(&balanceValue)
		if err != nil {
			return models.Balance{}, fmt.Errorf("get balance query row failed: %w", err)
		}
	} else {
		return models.Balance{}, fmt.Errorf("user_id %d: %w", userId, errors.UnknownUserIdError)
	}

	balanceDecimal, balErr := decimal.NewFromString(balanceValue)
	if balErr != nil {
		return models.Balance{}, fmt.Errorf("cannot get decimal balance from string %v", balanceValue)
	}
	balance := models.Balance{UserId: userId, Value: balanceDecimal}
	return balance, nil
}
