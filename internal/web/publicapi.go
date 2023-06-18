package web

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/sgladkov/tortuga/internal/logger"
	"go.uber.org/zap"
)

func mock(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func ConfigInfo(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func UserList(w http.ResponseWriter, _ *http.Request) {
	users, err := storage.GetUserList()
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

func UserInfo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := storage.GetUser(id)
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

func UserHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	projects, err := storage.GetUserProjects(id)
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

func ProjectList(w http.ResponseWriter, _ *http.Request) {
	projects, err := storage.GetUserList()
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

func ProjectInfo(w http.ResponseWriter, r *http.Request) {
	strId := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(strId, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	project, err := storage.GetProject(id)
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
