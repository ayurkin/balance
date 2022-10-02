package application

import (
	"balance/internal/adapters/http"
	"balance/internal/config"
	"context"
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
	app.logger.Sugar().Info(appConfig.HttpPort)
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
