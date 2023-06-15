package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/sgladkov/tortuga/internal/logger"
	"github.com/sgladkov/tortuga/internal/utils"
	"github.com/sgladkov/tortuga/internal/web"
	"go.uber.org/zap"
)

var db *sql.DB

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

	logger.Log.Info("Starting server", zap.String("address", config.Endpoint))
	err = http.ListenAndServe(config.Endpoint, web.TortugaRouter())
	if err != nil {
		logger.Log.Fatal("failed to start server", zap.Error(err))
	}
}
