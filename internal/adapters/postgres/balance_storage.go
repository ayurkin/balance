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

func (db *Database) AddIncome(ctx context.Context, transaction models.Transaction) error {
	tx, err := db.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx failed: %v", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO balance.history
				(from_id, to_id, amount, occurred_at, description)
			VALUES
				($1, $2, $3, $4, $5)`,
		0, transaction.ToId, transaction.Amount,
		transaction.Time, transaction.Description)
	if err != nil {
		return fmt.Errorf("query exec failed: %v", err)
	}

	var isUserIdExist bool

	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT user_id FROM balance.balance WHERE user_id = $1) AS exists",
		transaction.ToId).Scan(&isUserIdExist)

	if err != nil {
		return fmt.Errorf("query row failed: %w", err)
	}
	if isUserIdExist {
		_, err = tx.Exec(ctx,
			"UPDATE balance.balance SET balance = balance + $1 WHERE user_id = $2",
			transaction.Amount, transaction.ToId)
		if err != nil {
			return fmt.Errorf("query exec failed: %v", err)
		}
	} else {
		_, err = tx.Exec(ctx,
			"INSERT INTO balance.balance (user_id, balance) VALUES($1, $2)",
			transaction.ToId, transaction.Amount)
		if err != nil {
			return fmt.Errorf("query exec failed: %v", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx commit failed failed: %v", err)
	}
	return nil
}

func (db *Database) AddExpense(ctx context.Context, transaction models.Transaction) error {
	tx, err := db.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx failed: %v", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO balance.history
				(from_id, to_id, amount, occurred_at, description)
			VALUES
				($1, $2, $3, $4, $5)`,
		transaction.FromId, 0, transaction.Amount,
		transaction.Time, transaction.Description)
	if err != nil {
		return fmt.Errorf("query exec failed: %v", err)
	}

	var isUserIdExist bool

	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT user_id FROM balance.balance WHERE user_id = $1) AS exists",
		transaction.FromId).Scan(&isUserIdExist)

	if err != nil {
		return fmt.Errorf("query row failed: %w", err)
	}
	if isUserIdExist {
		_, err = tx.Exec(ctx,
			"UPDATE balance.balance SET balance = balance - $1 WHERE user_id = $2",
			transaction.Amount, transaction.FromId)
		if err != nil {

			if errPq, ok := err.(*pgconn.PgError); ok {
				if errPq.Code == pgerrcode.CheckViolation {
					return errors.NotEnoughUserBalance{UserError: errors.UserError{UserId: transaction.FromId}}
				}
			}

			return fmt.Errorf("query exec failed: %v", err)
		}
	} else {
		return errors.UnknownUserIdError{UserError: errors.UserError{UserId: transaction.FromId}}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx commit failed failed: %v", err)
	}
	return nil
}

func (db *Database) DoTransfer(ctx context.Context, transaction models.Transaction) error {
	tx, err := db.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx failed: %v", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO balance.history
				(from_id, to_id, amount, occurred_at, description)
			VALUES
				($1, $2, $3, $4, $5)`,
		transaction.FromId, transaction.ToId, transaction.Amount,
		transaction.Time, transaction.Description)
	if err != nil {
		return fmt.Errorf("query exec failed: %v", err)
	}

	var isUserIdFromExist bool

	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT user_id FROM balance.balance WHERE user_id = $1) AS exists",
		transaction.FromId).Scan(&isUserIdFromExist)

	if err != nil {
		return fmt.Errorf("query row failed: %w", err)
	}
	if isUserIdFromExist {
		_, err = tx.Exec(ctx,
			"UPDATE balance.balance SET balance = balance - $1 WHERE user_id = $2",
			transaction.Amount, transaction.FromId)
		if err != nil {

			if errPq, ok := err.(*pgconn.PgError); ok {
				if errPq.Code == pgerrcode.CheckViolation {
					return errors.NotEnoughUserBalance{UserError: errors.UserError{UserId: transaction.FromId}}
				}
			}

			return fmt.Errorf("query exec failed: %v", err)
		}
	} else {
		return errors.UnknownUserIdError{UserError: errors.UserError{UserId: transaction.FromId}}
	}

	var isUserIdToExist bool

	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT user_id FROM balance.balance WHERE user_id = $1) AS exists",
		transaction.ToId).Scan(&isUserIdToExist)

	if err != nil {
		return fmt.Errorf("query row failed: %w", err)
	}
	if isUserIdToExist {
		_, err = tx.Exec(ctx,
			"UPDATE balance.balance SET balance = balance + $1 WHERE user_id = $2",
			transaction.Amount, transaction.ToId)
		if err != nil {
			return fmt.Errorf("query exec failed: %v", err)
		}
	} else {
		_, err = tx.Exec(ctx,
			"INSERT INTO balance.balance (user_id, balance) VALUES($1, $2)",
			transaction.ToId, transaction.Amount)
		if err != nil {
			return fmt.Errorf("query exec failed: %v", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx commit failed failed: %v", err)
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
		return models.Balance{}, fmt.Errorf("query row failed: %w", err)
	}
	if isUserIdExist {
		err = db.DB.QueryRow(ctx,
			"SELECT balance FROM balance.balance WHERE user_id = $1", userId).Scan(&balanceValue)
		if err != nil {
			return models.Balance{}, fmt.Errorf("query row failed: %w", err)
		}
	} else {
		return models.Balance{}, errors.UnknownUserIdError{UserError: errors.UserError{UserId: userId}}
	}

	balanceDecimal, balErr := decimal.NewFromString(balanceValue)
	if balErr != nil {
		return models.Balance{}, fmt.Errorf("cannot get decimal balance from string %s", balanceValue)
	}
	balance := models.Balance{UserId: userId, Balance: balanceDecimal}
	return balance, nil
}
