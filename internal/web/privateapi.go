package web

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"net/http"
	"strconv"

	"github.com/sgladkov/tortuga/internal/logger"
	"github.com/sgladkov/tortuga/internal/models"

	"go.uber.org/zap"
)

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
	caller := r.Header.Get("TRTG-Address")
	if len(caller) == 0 {
		logger.Log.Warn("no caller in headers")
		http.Error(w, "no public key in headers", http.StatusBadRequest)
		return
	}

	logger.Log.Info("createProject", zap.String("user", caller), zap.Any("project", project))
	id, err := marketplace.CreateProject(r.Context(), caller, project.Title, project.Description, project.Tags, project.Owner, project.Deadline, project.Price)
	if err != nil {
		logger.Log.Warn("Failed to create project in marketplace", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to create project in marketplace [%s]", err), http.StatusBadRequest)
		return
	}
	project.Id = id
	project.Owner = caller
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
	caller := r.Header.Get("TRTG-Address")
	if len(caller) == 0 {
		logger.Log.Warn("no caller in headers")
		http.Error(w, "no public key in headers", http.StatusBadRequest)
		return
	}

	logger.Log.Info("updateProject", zap.String("user", caller), zap.Any("project", project))
	err = marketplace.UpdateProject(r.Context(), caller, project.Id, project.Title, project.Description, project.Tags, project.Deadline, project.Price)
	if err != nil {
		logger.Log.Warn("Failed to update project in marketplace", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update project in marketplace [%s]", err), http.StatusBadGateway)
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
	caller := r.Header.Get("TRTG-Address")
	if len(caller) == 0 {
		logger.Log.Warn("no caller in headers")
		http.Error(w, "no public key in headers", http.StatusBadRequest)
		return
	}

	logger.Log.Info("deleteProject", zap.String("user", caller), zap.Any("projectId", data.Id))
	err = marketplace.DeleteProject(r.Context(), caller, data.Id)
	if err != nil {
		logger.Log.Warn("Failed to update project in marketplace", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update project in marketplace [%s]", err), http.StatusBadGateway)
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

func cancelProject(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		logger.Log.Warn("invalid project id")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Log.Info("cancelProject", zap.String("user", caller), zap.Any("projectId", id))
	err = marketplace.CancelProject(r.Context(), caller, id)
	if err != nil {
		logger.Log.Warn("Failed to update project in marketplace", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update project in marketplace [%s]", err), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&id)
	if err != nil {
		logger.Log.Warn("Failed to write info to body", zap.Error(err))
		return
	}
}

func readyProject(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		logger.Log.Warn("invalid project id")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Log.Info("readyProject", zap.String("user", caller), zap.Any("projectId", id))
	err = marketplace.SetProjectReady(r.Context(), caller, id)
	if err != nil {
		logger.Log.Warn("Failed to update project in marketplace", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update project in marketplace [%s]", err), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&id)
	if err != nil {
		logger.Log.Warn("Failed to write info to body", zap.Error(err))
		return
	}
}

func acceptProject(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		logger.Log.Warn("invalid project id")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Log.Info("acceptProject", zap.String("user", caller), zap.Any("projectId", id))
	err = marketplace.AcceptProject(r.Context(), caller, id)
	if err != nil {
		logger.Log.Warn("Failed to update project in marketplace", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update project in marketplace [%s]", err), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&id)
	if err != nil {
		logger.Log.Warn("Failed to write info to body", zap.Error(err))
		return
	}
}

func createBid(w http.ResponseWriter, r *http.Request) {
	if !ContainsHeaderValue(r, "Content-Type", "application/json") {
		contentType := r.Header.Get("Content-Type")
		logger.Log.Warn("Wrong Content-Type header", zap.String("Content-Type", contentType))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", contentType), http.StatusBadRequest)
		return
	}
	var bid models.Bid
	err := json.NewDecoder(r.Body).Decode(&bid)
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
	projectId, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		logger.Log.Warn("invalid project id")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Log.Info("createBid", zap.String("user", owner), zap.Uint64("project", projectId),
		zap.Any("bid", bid))
	id, err := marketplace.CreateBid(r.Context(), owner, projectId, owner, bid.Price, bid.Deadline, bid.Message)
	if err != nil {
		logger.Log.Warn("Failed to create bid in marketplace", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to create bid in marketplace [%s]", err), http.StatusBadRequest)
		return
	}
	bid.Id = id
	bid.User = owner
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&bid)
	if err != nil {
		logger.Log.Warn("Failed to write info to body", zap.Error(err))
		return
	}
}

func acceptBid(w http.ResponseWriter, r *http.Request) {
	if !ContainsHeaderValue(r, "Content-Type", "application/json") {
		contentType := r.Header.Get("Content-Type")
		logger.Log.Warn("Wrong Content-Type header", zap.String("Content-Type", contentType))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", contentType), http.StatusBadRequest)
		return
	}
	owner := r.Header.Get("TRTG-Address")
	if len(owner) == 0 {
		logger.Log.Warn("no owner in headers")
		http.Error(w, "no public key in headers", http.StatusBadRequest)
		return
	}
	bidId, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		logger.Log.Warn("invalid project id")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logger.Log.Info("acceptBid", zap.String("user", owner), zap.Uint64("bid", bidId))
	projectId, err := marketplace.AcceptBid(r.Context(), owner, bidId)
	if err != nil {
		logger.Log.Warn("Failed to accept bid in marketplace", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to accept bid in marketplace [%s]", err), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	project, err := marketplace.GetProject(r.Context(), projectId)
	if err != nil {
		logger.Log.Warn("invalid project in bid")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(&project)
	if err != nil {
		logger.Log.Warn("Failed to write info to body", zap.Error(err))
		return
	}
}

func updateBid(w http.ResponseWriter, r *http.Request) {
	if !ContainsHeaderValue(r, "Content-Type", "application/json") {
		contentType := r.Header.Get("Content-Type")
		logger.Log.Warn("Wrong Content-Type header", zap.String("Content-Type", contentType))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", contentType), http.StatusBadRequest)
		return
	}
	bidId, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		logger.Log.Warn("invalid project id")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var bid models.Bid
	err = json.NewDecoder(r.Body).Decode(&bid)
	if err != nil {
		logger.Log.Warn("Failed to decode JSON to Bid", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to decode JSON to Bid [%s]", err), http.StatusBadRequest)
		return
	}
	bid.Id = bidId
	owner := r.Header.Get("TRTG-Address")
	if len(owner) == 0 {
		logger.Log.Warn("no owner in headers")
		http.Error(w, "no public key in headers", http.StatusBadRequest)
		return
	}

	logger.Log.Info("updateBid", zap.String("user", owner), zap.Any("bid", bid))
	err = marketplace.UpdateBid(r.Context(), owner, bid.Id, bid.Price, bid.Deadline, bid.Message)
	if err != nil {
		logger.Log.Warn("Failed to update bid in marketplace", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update bid in marketplace [%s]", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&bid)
	if err != nil {
		logger.Log.Warn("Failed to write info to body", zap.Error(err))
		return
	}
}

func deleteBid(w http.ResponseWriter, r *http.Request) {
	if !ContainsHeaderValue(r, "Content-Type", "application/json") {
		contentType := r.Header.Get("Content-Type")
		logger.Log.Warn("Wrong Content-Type header", zap.String("Content-Type", contentType))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", contentType), http.StatusBadRequest)
		return
	}
	bidId, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		logger.Log.Warn("invalid project id")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	caller := r.Header.Get("TRTG-Address")
	if len(caller) == 0 {
		logger.Log.Warn("no caller in headers")
		http.Error(w, "no public key in headers", http.StatusBadRequest)
		return
	}

	logger.Log.Info("deleteBid", zap.String("user", caller), zap.Any("bidId", bidId))
	err = marketplace.DeleteBid(r.Context(), caller, bidId)
	if err != nil {
		logger.Log.Warn("Failed to delete bid in marketplace", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to delete bid in marketplace [%s]", err), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&bidId)
	if err != nil {
		logger.Log.Warn("Failed to write info to body", zap.Error(err))
		return
	}
}
