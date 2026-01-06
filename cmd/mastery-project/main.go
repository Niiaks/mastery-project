package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"mastery-project/internal/config"
	"mastery-project/internal/database"
	"mastery-project/internal/handler"
	"mastery-project/internal/middleware"
	"mastery-project/internal/repository"
	"mastery-project/internal/router"
	"mastery-project/internal/server"
	"mastery-project/internal/service"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	//load env variables on startup
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	//start http server
	srv, err := server.NewServer(cfg)
	if err != nil {
		panic(err)
	}

	if err := database.RunMigrations(cfg); err != nil {
		log.Fatal(err)
	}

	repos := repository.NewRepository(srv.Db.Pool)

	services, serviceErr := service.NewServices(repos)
	if serviceErr != nil {
		panic(serviceErr)
	}
	//setup handlers
	handlers := handler.NewHandlers(cfg, services)

	authMW := middleware.NewAuthMiddleware(repos.Session)
	r := router.NewRouter(handlers, authMW)

	srv.SetupHttpServer(r)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	//start server
	go func() {
		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error(err.Error())
			log.Fatal(err.Error())
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err = srv.Shutdown(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error(err.Error())
		log.Fatal(err.Error())
	}
	stop()
	cancel()

	slog.Info("server shutdown properly")
}
