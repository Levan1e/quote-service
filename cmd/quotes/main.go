package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	v1 "quote-service/internal/api/v1"
	repoPostgres "quote-service/internal/repository/postgres"
	quoteshttp "quote-service/pkg/http"
	"quote-service/pkg/logger"
	"quote-service/pkg/postgres"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// Инициализация конфигурации
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		logger.Fatalf("Ошибка чтения конфигурации: %v", err)
	}

	// Инициализация логгера
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Ошибка инициализации логгера: %v", err))
	}
	defer logger.Sync()

	// Подключение к базе данных
	db, err := postgres.NewPostgres(postgres.Options{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		DBName:   viper.GetString("database.dbname"),
	})
	if err != nil {
		logger.Fatal("Ошибка подключения к БД", zap.Error(err))
	}
	defer db.Close()

	// Инициализация репозитория
	storage := repoPostgres.NewStorage(db.DB)

	// Инициализация роутера
	r := chi.NewRouter()

	handler := v1.NewHandler(storage, logger)
	r.Mount("/quotes", handler.Routes())

	// Создание HTTP-сервера
	port := viper.GetInt("server.port")
	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Запуск сервера
	go func() {
		logger.Info("Сервер запущен", zap.Int("port", port))
		if err := quoteshttp.ListenAndServe(addr, r); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Ошибка запуска сервера", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop // Ожидание сигнала (например, Ctrl+C)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Завершение работы сервера
	logger.Info("Завершение работы сервера...")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Ошибка при завершении работы сервера", zap.Error(err))
	}

	// Закрытие соединения с базой данных
	logger.Info("Закрытие соединения с базой данных...")
	db.Close()

	logger.Info("Сервер успешно завершил работу")
}
