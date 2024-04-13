package main

import (
	"avito-backend-assignment/internal/handlers"
	"avito-backend-assignment/internal/middleware"
	"database/sql"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {

	connStr := "user=avito password=avito dbname=avito_banner_db host=postgres port=5432 sslmode=disable"
	MemcachedAddresses := []string{"memcached:11211"}

	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

	var db *sql.DB
	var err error
	maxAttempts := 10

	// так как postgres в docker compose может быть не готов принимать подключение проверяем его 5 минут
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err = sql.Open("postgres", connStr)
		// Проверка подключения к базе данных
		err = db.Ping()

		if err == nil {
			break
		}

		logger.Infof("Failed to connect to PostgreSQL (attempt %d): %v\n", attempt, err)
		if attempt == maxAttempts {
			panic(fmt.Errorf("failed to connect to PostgreSQL after %d attempts: %v", maxAttempts, err))
		}

		logger.Info("Retrying in 30 seconds...")
		time.Sleep(30 * time.Second)
	}

	memcacheClient := memcache.New(MemcachedAddresses...)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = memcacheClient.Ping()
		if err == nil {
			break
		}
		logger.Infof("Failed to connect to memcached (attempt %d): %v\n", attempt, err)
		if attempt == maxAttempts {
			panic(fmt.Errorf("failed to connect to memcached after %d attempts: %v", maxAttempts, err))
		}

		logger.Info("Retrying in 15 seconds...")
		time.Sleep(15 * time.Second)
	}

	err = memcacheClient.DeleteAll()
	if err != nil {
		panic(err)
	}

	usersHandler := &handlers.UsersHandler{
		Logger: logger,
		DB:     db,
		MC:     memcacheClient,
	}

	adminsHandler := &handlers.AdminsHandler{
		Logger: logger,
		DB:     db,
		MC:     memcacheClient,
	}

	r := mux.NewRouter()
	r.HandleFunc("/user_banner", usersHandler.Banner).Methods("GET")
	r.HandleFunc("/banner", adminsHandler.NewBanner).Methods("POST")
	r.HandleFunc("/banner/{id}", adminsHandler.DeleteBanner).Methods("DELETE")
	r.HandleFunc("/banner/{id}", adminsHandler.UpdateBanner).Methods("PATCH")
	r.HandleFunc("/banner", adminsHandler.Banner).Methods("GET")
	muxRouter := middleware.Auth(r)
	muxRouter = middleware.AccessLog(logger, muxRouter)
	muxRouter = middleware.Panic(muxRouter)

	addr := ":8080"
	logger.Infow("starting server",
		"type", "START",
		"addr", addr,
	)
	err = http.ListenAndServe(addr, muxRouter)
	if err != nil {
		logger.Panic(err.Error())
	}

}
