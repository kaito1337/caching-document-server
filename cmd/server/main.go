package main

import (
	"context"
	"document-server/internal/api"
	"document-server/internal/api/controller"
	"document-server/internal/api/middleware"
	"document-server/internal/cache"
	"document-server/internal/config"
	"document-server/internal/infrastructure/database/postgres"
	"document-server/internal/logger"
	"document-server/internal/service"
	document "document-server/internal/storage/document"
	token "document-server/internal/storage/token"
	user "document-server/internal/storage/user"

	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.LoadConfig("configs/config.json")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := logger.New(cfg.Log.Level)
	slog.SetDefault(logger)

	logger.Info("configuration loaded", slog.String("addr", cfg.Server.Address))

	if _, err := os.Stat(cfg.FileStorage.Path); os.IsNotExist(err) {
		os.MkdirAll(cfg.FileStorage.Path, 0755)
		slog.Info("created file storage directory", slog.String("path", cfg.FileStorage.Path))
	}

	db, err := postgres.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	log.Println("Database connected")

	docStorage := document.NewDocumentStorage(db)
	userStorage := user.NewUserStorage(db)
	tokenStorage := token.NewTokenStorage(db)

	inMemoryCache := cache.NewInMemoryCache(cfg.CacheConfig)

	docService := service.NewDocumentService(userStorage, docStorage, tokenStorage, cfg.FileStorage.Path, inMemoryCache)
	authService := service.NewUserService(userStorage, tokenStorage, cfg.AdminToken)

	userAuthMiddleware := middleware.NewUserAuthMiddleware(authService)

	router, err := api.NewRouter(userAuthMiddleware)

	userController := controller.NewUserController(authService)
	docsController := controller.NewDocumentController(docService)

	router.SetUserRoutes(userController)
	router.SetDocsRoutes(docsController)

	srv := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("starting server", slog.String("addr", cfg.Server.Address))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-stop
	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown gracefully", slog.String("error", err.Error()))
	} else {
		logger.Info("server stopped gracefully")
	}
}
