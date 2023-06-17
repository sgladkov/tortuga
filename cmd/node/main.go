package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/sgladkov/tortuga/internal/logger"
	"github.com/sgladkov/tortuga/internal/models"
	storage2 "github.com/sgladkov/tortuga/internal/storage"
	"github.com/sgladkov/tortuga/internal/utils"
	"github.com/sgladkov/tortuga/internal/web"
	"go.uber.org/zap"
)

func main() {
	config := utils.Config{}
	err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	err = logger.Initialize(config.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	var storage storage2.Storage
	if len(config.DatabaseDSN) > 0 {
		db, err := sql.Open("postgres", config.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("failed to open database", zap.Error(err))
		}
		defer func() {
			err := db.Close()
			if err != nil {
				logger.Log.Warn("failed to close database", zap.Error(err))
			}
		}()
		storage, err = storage2.NewPgStorage(db)
		if err != nil {
			logger.Log.Fatal("failed to init storage", zap.Error(err))
		}
	} else {
		storage = storage2.NewTestStorage([]models.User{}, []models.Project{}, []models.Bid{}, []models.Rate{})
	}
	logger.Log.Info("Starting server", zap.String("address", config.Endpoint))
	err = http.ListenAndServe(config.Endpoint, web.TortugaRouter(storage))
	if err != nil {
		logger.Log.Fatal("failed to start server", zap.Error(err))
	}
}
