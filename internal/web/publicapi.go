package web

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/sgladkov/tortuga/internal/logger"

	"go.uber.org/zap"
)

func configInfo(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	data := struct {
		ExchangeWallet string
	}{address}
	err := json.NewEncoder(w).Encode(&data)
	if err != nil {
		logger.Log.Warn("Failed to write info to body", zap.Error(err))
		return
	}
}

func userList(w http.ResponseWriter, _ *http.Request) {
	users, err := marketplace.GetUserList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&users)
	if err != nil {
		logger.Log.Warn("Failed to write user list to body", zap.Error(err))
		return
	}
}

func userInfo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := marketplace.GetUser(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&user)
	if err != nil {
		logger.Log.Warn("Failed to write user to body", zap.Error(err))
		return
	}
}

func userHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	projects, err := marketplace.GetUserProjects(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&projects)
	if err != nil {
		logger.Log.Warn("Failed to write user history to body", zap.Error(err))
		return
	}
}

func projectList(w http.ResponseWriter, _ *http.Request) {
	projects, err := marketplace.GetProjectList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&projects)
	if err != nil {
		logger.Log.Warn("Failed to write project list to body", zap.Error(err))
		return
	}
}

func projectInfo(w http.ResponseWriter, r *http.Request) {
	strId := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(strId, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	project, err := marketplace.GetProject(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&project)
	if err != nil {
		logger.Log.Warn("Failed to write project to body", zap.Error(err))
		return
	}
}

func projectBids(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bids, err := marketplace.GetProjectBids(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&bids)
	if err != nil {
		logger.Log.Warn("Failed to write user history to body", zap.Error(err))
		return
	}
}

func bidInfo(w http.ResponseWriter, r *http.Request) {
	strId := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(strId, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bid, err := marketplace.GetBid(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&bid)
	if err != nil {
		logger.Log.Warn("Failed to write bid to body", zap.Error(err))
		return
	}
}
