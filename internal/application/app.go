package application

import (
	"balance/internal/adapters/http"
	"balance/internal/adapters/postgres"
	"balance/internal/config"
	"balance/internal/domain/models"
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"time"
)

type App struct {
	logger     *zap.Logger
	httpServer *http.Server
}

func Start(ctx context.Context, app *App) {
	logger, _ := zap.NewProduction()
	app.logger = logger

	appConfig, err := config.NewConfig()
	if err != nil {
		logger.Sugar().Fatalf("create config failed: %v", err)
	}

	pgconn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		appConfig.PostgresUser, appConfig.PostgresPassword, appConfig.PostgresHost,
		appConfig.PostgresPort, appConfig.PostgresDb)

	db, err := postgres.New(ctx, pgconn)
	if err != nil {
		logger.Sugar().Fatalf("db init failed: %v", err)
	}

	balance, BalErr := db.GetBalance(ctx, 1)
	app.logger.Sugar().Info("1 GetBalance user_id 1", balance, BalErr)
	balance, BalErr = db.GetBalance(ctx, 2)
	app.logger.Sugar().Info("2 GetBalance user_id 2", balance, BalErr)

	amount, _ := decimal.NewFromString("100.00")
	trans := models.Transaction{
		FromId:      0,
		ToId:        1,
		Amount:      amount,
		Time:        time.Now(),
		Description: "Salary",
	}

	err = db.AddIncome(ctx, trans)
	app.logger.Sugar().Info("3 Add income user_id 1", err)

	balance, BalErr = db.GetBalance(ctx, 1)
	app.logger.Sugar().Info("4 GetBalance user_id 1", balance, BalErr)

	trans.FromId = 1
	trans.ToId = 2

	err = db.DoTransfer(ctx, trans)
	app.logger.Sugar().Info("4 Transfer from user_id 1 to 2", err)
	balance, BalErr = db.GetBalance(ctx, 1)
	app.logger.Sugar().Info("6 GetBalance user_id 1", balance, BalErr)
	balance, BalErr = db.GetBalance(ctx, 2)
	app.logger.Sugar().Info("7 GetBalance user_id 2", balance, BalErr)

	trans.FromId = 2
	trans.ToId = 0

	err = db.AddExpense(ctx, trans)
	app.logger.Sugar().Info("8 Add expense user_id 2", err)

	balance, BalErr = db.GetBalance(ctx, 2)
	app.logger.Sugar().Info("9 GetBalance user_id 2", balance, BalErr)

	app.httpServer = http.New(logger.Sugar())

	go func() {
		err := app.httpServer.Start(appConfig.HttpPort)
		if err != nil {
			logger.Sugar().Fatalf("http server failed: %v", err)
		}
	}()

	app.logger.Sugar().Info("application has started")
}

func Stop(app *App) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.httpServer.Stop(ctx)
	if err != nil {
		app.logger.Sugar().Errorf("stop http server failed: %v", err)
	}

	app.logger.Sugar().Info("app has stopped")
}
