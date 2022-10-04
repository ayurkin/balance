package balance

import (
	e "balance/internal/domain/errors"
	"balance/internal/domain/models"
	"balance/internal/ports"
	"context"
	"errors"
	"go.uber.org/zap"
)

type Service struct {
	db     ports.BalanceStoragePort
	logger *zap.SugaredLogger
}

func New(db ports.BalanceStoragePort, logger *zap.SugaredLogger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

func (s *Service) AddIncome(ctx context.Context, transaction models.BalanceWithDesc) error {
	err := s.db.AddIncome(ctx, transaction)

	if err != nil {
		s.logger.Errorf("add income fail: %v", err)
		return e.DatabaseError
	}
	return nil
}

func (s *Service) AddExpense(ctx context.Context, transaction models.BalanceWithDesc) error {
	err := s.db.AddExpense(ctx, transaction)

	if err != nil {
		s.logger.Errorf("add expense fail: %v", err)
		if errors.Is(err, e.UnknownUserIdError) || errors.Is(err, e.NotEnoughUserBalanceError) {
			return err
		}
		return e.DatabaseError
	}
	return nil
}

func (s *Service) DoTransfer(ctx context.Context, transaction models.Transaction) error {
	err := s.db.DoTransfer(ctx, transaction)

	if err != nil {
		s.logger.Errorf("transfer fail: %v", err)
		if errors.Is(err, e.UnknownUserIdError) || errors.Is(err, e.NotEnoughUserBalanceError) {
			return err
		}
		return e.DatabaseError
	}
	return nil
}

func (s *Service) GetBalance(ctx context.Context, userId int64) (models.Balance, error) {
	balance, err := s.db.GetBalance(ctx, userId)

	if err != nil {
		s.logger.Errorf("get balance fail: %v", err)
		if errors.Is(err, e.UnknownUserIdError) {
			return models.Balance{}, err
		}
		return models.Balance{}, e.DatabaseError
	}
	return balance, nil
}
