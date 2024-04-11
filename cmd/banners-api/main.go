package main

import (
	"avito-backend-assignment/internal/handlers"
	"avito-backend-assignment/internal/middleware"
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	fmt.Println(time.Now().String())
	connStr := "user=avito password=avito dbname=avito_banner_db host=localhost port=5433 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	// Проверка подключения к базе данных
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

	usersHandler := &handlers.UsersHandler{
		Logger: logger,
		DB:     db,
	}

	adminsHandler := &handlers.AdminsHandler{
		Logger: logger,
		DB:     db,
	}

	r := mux.NewRouter()
	r.HandleFunc("/user_banner", usersHandler.Banner).Methods("GET")
	r.HandleFunc("/banner", adminsHandler.NewBanner).Methods("POST")

	//mux := middleware.Auth(sm, r)
	//mux = middleware.AccessLog(logger, mux)
	muxRouter := middleware.Panic(r)

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
