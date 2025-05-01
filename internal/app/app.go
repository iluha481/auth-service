package app

import (
	"log/slog"
	"time"

	grpcapp "sso/internal/app/grpc"
	"sso/internal/services/auth"
	"sso/internal/storage/postgresql"
)

type App struct {
	GRPCServer *grpcapp.App
	Storage    *postgresql.Storage
}

func New(
	log *slog.Logger,
	grpcPort int,
	connection string,
	tokenTTL time.Duration,
) *App {
	storage, err := postgresql.New(connection)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
		Storage:    storage,
	}
}
