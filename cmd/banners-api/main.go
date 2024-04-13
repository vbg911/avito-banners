package main

import (
	"avito-backend-assignment/internal/handlers"
	"avito-backend-assignment/internal/middleware"
	"database/sql"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"net/http"
)

// bombardier -c 100 -n 100000 "http://127.0.0.1:8080/user_banner?tag_id=1&feature_id=1&use_last_revision=0" -H "token:user"
func main() {
	connStr := "user=avito password=avito dbname=avito_banner_db host=localhost port=5433 sslmode=disable"
	MemcachedAddresses := []string{"127.0.0.1:11211"}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	// Проверка подключения к базе данных
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	memcacheClient := memcache.New(MemcachedAddresses...)
	err = memcacheClient.Ping()
	if err != nil {
		panic(err)
	}
	err = memcacheClient.DeleteAll()
	if err != nil {
		panic(err)
	}

	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

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
