package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	coinbaseInfra "CourseWork/infrastructure/coinbase"
	supabaseDB "CourseWork/infrastructure/db/supabase"
	"CourseWork/infrastructure/logger"
	tgInfra "CourseWork/infrastructure/telegram"
	tgDelivery "CourseWork/internal/delivery/telegram"
	currencyUsecase "github.com/4rt3mio/cryptoCore/usecase/currency"
	notifUsecase "github.com/4rt3mio/cryptoCore/usecase/notification"
	subUsecase "github.com/4rt3mio/cryptoCore/usecase/subscription"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {
	zapLogger, err := logger.NewZapLogger()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	log := zapLogger
	log.Info("Starting application...")

	db, err := sqlx.Connect("pgx", os.Getenv("DATABASE_URL"))
	db.SetConnMaxLifetime(30 * time.Second)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	if err != nil {
		log.Error("DB connect failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("Connected to database")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Error("Error checking database connection", "err", err)
		os.Exit(1)
	}

	tgClient, err := tgInfra.NewClient()
	if err != nil {
		log.Error("Telegram client init failed", "err", err)
		os.Exit(1)
	}
	log.Info("Telegram client initialized")

	cbClient := coinbaseInfra.NewClient(log)
	cryptoRepo := coinbaseInfra.NewCryptoRepositoryAdapter(cbClient)

	subRepo := supabaseDB.NewSubscriptionRepository(db, log)
	subMgr := subUsecase.NewManager(subRepo)
	monitorSvc := subUsecase.NewMonitorService(cryptoRepo, 30*time.Second)
	notifier := notifUsecase.NewTelegramNotifier(tgClient)
	currMgr := currencyUsecase.NewManager(coinbaseInfra.NewCurrencyRepositoryAdapter(cbClient))
	currencyAnalytics := currencyUsecase.NewAnalytics(cryptoRepo)

	controller := tgDelivery.NewTelegramController(
		tgClient,
		currencyAnalytics,
		subMgr,
		monitorSvc,
		notifier,
		currMgr,
	)

	go controller.Start()
	log.Info("Telegram bot launched")
	fmt.Println("Telegram-бот запущен. Нажмите Ctrl+C для остановки.")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Остановка бота…")
	log.Info("Telegram bot stopped")
}
