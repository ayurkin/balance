package application

import (
	"balance/internal/adapters/http"
	"balance/internal/adapters/postgres"
	"balance/internal/config"
	"balance/internal/domain/balance"
	"balance/internal/utils"
	"context"
	"fmt"
	_ "github.com/lib/pq"
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

	err = utils.ApplyMigrations(pgconn, "db/changelog/")
	if err != nil {
		logger.Sugar().Fatalf("migrations failed: %v", err)
	}

	balanceS := balance.New(db, logger.Sugar())

	app.httpServer = http.New(balanceS, logger.Sugar())

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
