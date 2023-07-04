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
		logger.Log.Warn("Failed to decode JSON to User", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to decode JSON to User [%s]", err), http.StatusBadRequest)
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
	err = json.NewEncoder(w).Encode(&user)
	if err != nil {
		logger.Log.Warn("Failed to write info to body", zap.Error(err))
		return
	}
}

func createProject(w http.ResponseWriter, r *http.Request) {
	if !ContainsHeaderValue(r, "Content-Type", "application/json") {
		contentType := r.Header.Get("Content-Type")
		logger.Log.Warn("Wrong Content-Type header", zap.String("Content-Type", contentType))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", contentType), http.StatusBadRequest)
		return
	}
	var project models.Project
	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		logger.Log.Warn("Failed to decode JSON to Project", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to decode JSON to Project [%s]", err), http.StatusBadRequest)
		return
	}
	owner := r.Header.Get("TRTG-Address")
	if len(owner) == 0 {
		logger.Log.Warn("no owner in headers")
		http.Error(w, "no public key in headers", http.StatusBadRequest)
		return
	}

	logger.Log.Info("createProject", zap.String("user", owner), zap.Any("project", project))
	id, err := storage.CreateProject(project.Title, project.Description, project.Tags, owner, project.Deadline, project.Price)
	if err != nil {
		logger.Log.Warn("Failed to create project in storage", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to create project in storage [%s]", err), http.StatusBadRequest)
		return
	}
	project.Id = id
	project.Owner = owner
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&project)
	if err != nil {
		logger.Log.Warn("Failed to write info to body", zap.Error(err))
		return
	}
}

func updateProject(w http.ResponseWriter, r *http.Request) {
	if !ContainsHeaderValue(r, "Content-Type", "application/json") {
		contentType := r.Header.Get("Content-Type")
		logger.Log.Warn("Wrong Content-Type header", zap.String("Content-Type", contentType))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", contentType), http.StatusBadRequest)
		return
	}
	var project models.Project
	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		logger.Log.Warn("Failed to decode JSON to Project", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to decode JSON to Project [%s]", err), http.StatusBadRequest)
		return
	}
	owner := r.Header.Get("TRTG-Address")
	if len(owner) == 0 {
		logger.Log.Warn("no owner in headers")
		http.Error(w, "no public key in headers", http.StatusBadRequest)
		return
	}

	savedProject, err := storage.GetProject(project.Id)
	if err != nil {
		logger.Log.Warn("Failed to load saved project", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to load saved project [%s]", err), http.StatusBadRequest)
		return
	}

	if savedProject.Owner != owner {
		logger.Log.Warn("Invalid owner", zap.String("project owner", savedProject.Owner),
			zap.String("request sender", owner))
		http.Error(w, "Invalid project owner", http.StatusBadRequest)
		return
	}

	logger.Log.Info("updateProject", zap.String("user", owner), zap.Any("project", project))
	err = storage.UpdateProject(project.Id, project.Title, project.Description, project.Tags, project.Deadline, project.Price)
	if err != nil {
		logger.Log.Warn("Failed to update project in storage", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update project in storage [%s]", err), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&project)
	if err != nil {
		logger.Log.Warn("Failed to write info to body", zap.Error(err))
		return
	}
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	if !ContainsHeaderValue(r, "Content-Type", "application/json") {
		contentType := r.Header.Get("Content-Type")
		logger.Log.Warn("Wrong Content-Type header", zap.String("Content-Type", contentType))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", contentType), http.StatusBadRequest)
		return
	}
	var data struct {
		Id uint64
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		logger.Log.Warn("Failed to decode JSON", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to decode JSON [%s]", err), http.StatusBadRequest)
		return
	}
	owner := r.Header.Get("TRTG-Address")
	if len(owner) == 0 {
		logger.Log.Warn("no owner in headers")
		http.Error(w, "no public key in headers", http.StatusBadRequest)
		return
	}

	savedProject, err := storage.GetProject(data.Id)
	if err != nil {
		logger.Log.Warn("Failed to load project", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to load project [%s]", err), http.StatusBadRequest)
		return
	}

	if savedProject.Owner != owner {
		logger.Log.Warn("Invalid owner", zap.String("project owner", savedProject.Owner),
			zap.String("request sender", owner))
		http.Error(w, "Invalid project owner", http.StatusBadRequest)
		return
	}

	if savedProject.Status != models.Open {
		logger.Log.Warn("Invalid project status", zap.Uint("project status",
			uint(savedProject.Status)))
		http.Error(w, "Invalid project status", http.StatusBadRequest)
		return
	}

	logger.Log.Info("deleteProject", zap.String("user", owner), zap.Any("projectId", data.Id))
	err = storage.DeleteProject(data.Id)
	if err != nil {
		logger.Log.Warn("Failed to update project in storage", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update project in storage [%s]", err), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&data)
	if err != nil {
		logger.Log.Warn("Failed to write info to body", zap.Error(err))
		return
	}
}
