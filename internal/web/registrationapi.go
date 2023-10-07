package web

import (
	"encoding/json"
	"fmt"
	"github.com/sgladkov/tortuga/internal/logger"
	"github.com/sgladkov/tortuga/internal/models"
	"go.uber.org/zap"
	"net/http"
)

func register(w http.ResponseWriter, r *http.Request) {
	if !ContainsHeaderValue(r, "Content-Type", "application/json") {
		contentType := r.Header.Get("Content-Type")
		logger.Log.Warn("Wrong Content-Type header", zap.String("Content-Type", contentType))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", contentType), http.StatusBadRequest)
		return
	}

	caller := r.Header.Get("TRTG-Address")
	if len(caller) == 0 {
		logger.Log.Warn("no caller in headers")
		http.Error(w, "no public key in headers", http.StatusBadRequest)
		return
	}

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		logger.Log.Warn("Failed to decode JSON to User", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to decode JSON to User [%s]", err), http.StatusBadRequest)
		return
	}
	logger.Log.Info("register", zap.Any("user", user))
	err = marketplace.AddUser(r.Context(), caller, user)
	if err != nil {
		logger.Log.Warn("Failed to add user to marketplace", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to add user to marketplace [%s]", err), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&user)
	if err != nil {
		logger.Log.Warn("Failed to write info to body", zap.Error(err))
		return
	}
}


