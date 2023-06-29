package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sgladkov/tortuga/internal/logger"
	"github.com/sgladkov/tortuga/internal/models"

	"go.uber.org/zap"
)

func mock(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func register(w http.ResponseWriter, r *http.Request) {
	if !ContainsHeaderValue(r, "Content-Type", "application/json") {
		contentType := r.Header.Get("Content-Type")
		logger.Log.Warn("Wrong Content-Type header", zap.String("Content-Type", contentType))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", contentType), http.StatusBadRequest)
		return
	}
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		logger.Log.Warn("Failed to decode JSON to Metrics", zap.Error(err))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", err), http.StatusBadRequest)
		return
	}
	logger.Log.Info("register", zap.Any("user", user))
	err = storage.AddUser(&user)
	if err != nil {
		logger.Log.Warn("Failed to add user to storage", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to add user to storage [%s]", err), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
