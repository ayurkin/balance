package main

import (
	"balance/internal/application"
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	app := application.App{}

	go application.Start(ctx, &app)
	<-ctx.Done()
	application.Stop(&app)
}
