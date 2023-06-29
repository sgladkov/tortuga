package main

import (
	"log"
	"net/http"

	"github.com/sgladkov/tortuga/internal/logger"
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

	address, err := initAddress(&config)
	if err != nil {
		logger.Log.Fatal("failed to init exchange wallet address", zap.Error(err))
	}
	logger.Log.Info("Wallet address", zap.String("address", address))

	storage, err := initStorage(&config)
	if err != nil {
		logger.Log.Fatal("failed to init storage", zap.Error(err))
	}
	defer func() {
		err := storage.Close()
		if err != nil {
			logger.Log.Warn("failed to close storage", zap.Error(err))
		}
	}()

	logger.Log.Info("Starting server", zap.String("host", config.Endpoint))
	err = http.ListenAndServe(config.Endpoint, web.TortugaRouter(storage, address))
	if err != nil {
		logger.Log.Fatal("failed to start server", zap.Error(err))
	}
}
