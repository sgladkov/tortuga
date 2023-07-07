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

func cancelProject(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		logger.Log.Warn("invalid project id")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	savedProject, err := storage.GetProject(id)
	if err != nil {
		logger.Log.Warn("Failed to load project", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to load project [%s]", err), http.StatusBadRequest)
		return
	}

	if savedProject.Status != models.InWork {
		logger.Log.Warn("Invalid project status", zap.Uint8("status", uint8(savedProject.Status)))
		http.Error(w, "Invalid project status", http.StatusBadRequest)
		return
	}

	if savedProject.Owner != owner && savedProject.Contractor != owner {
		logger.Log.Warn("Invalid user", zap.String("project owner", savedProject.Owner),
			zap.String("project contractor", savedProject.Contractor),
			zap.String("request sender", owner))
		http.Error(w, "Invalid user", http.StatusForbidden)
		return
	}

	logger.Log.Info("cancelProject", zap.String("user", owner), zap.Any("projectId", id))
	err = storage.CancelProject(id)
	if err != nil {
		logger.Log.Warn("Failed to update project in storage", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update project in storage [%s]", err), http.StatusBadGateway)
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
	owner := r.Header.Get("TRTG-Address")
	if len(owner) == 0 {
		logger.Log.Warn("no owner in headers")
		http.Error(w, "no public key in headers", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		logger.Log.Warn("invalid project id")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	savedProject, err := storage.GetProject(id)
	if err != nil {
		logger.Log.Warn("Failed to load project", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to load project [%s]", err), http.StatusBadRequest)
		return
	}

	if savedProject.Status != models.InWork {
		logger.Log.Warn("Invalid project status", zap.Uint8("status", uint8(savedProject.Status)))
		http.Error(w, "Invalid project status", http.StatusBadRequest)
		return
	}

	if savedProject.Contractor != owner {
		logger.Log.Warn("Invalid user", zap.String("project owner", savedProject.Owner),
			zap.String("project contractor", savedProject.Contractor),
			zap.String("request sender", owner))
		http.Error(w, "Invalid user", http.StatusForbidden)
		return
	}

	logger.Log.Info("readyProject", zap.String("user", owner), zap.Any("projectId", id))
	err = storage.SetProjectReady(id)
	if err != nil {
		logger.Log.Warn("Failed to update project in storage", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update project in storage [%s]", err), http.StatusBadGateway)
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
	owner := r.Header.Get("TRTG-Address")
	if len(owner) == 0 {
		logger.Log.Warn("no owner in headers")
		http.Error(w, "no public key in headers", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		logger.Log.Warn("invalid project id")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	savedProject, err := storage.GetProject(id)
	if err != nil {
		logger.Log.Warn("Failed to load project", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to load project [%s]", err), http.StatusBadRequest)
		return
	}

	if savedProject.Status != models.InReview {
		logger.Log.Warn("Invalid project status", zap.Uint8("status", uint8(savedProject.Status)))
		http.Error(w, "Invalid project status", http.StatusBadRequest)
		return
	}

	if savedProject.Owner != owner {
		logger.Log.Warn("Invalid user", zap.String("project owner", savedProject.Owner),
			zap.String("project contractor", savedProject.Contractor),
			zap.String("request sender", owner))
		http.Error(w, "Invalid user", http.StatusForbidden)
		return
	}

	logger.Log.Info("acceptProject", zap.String("user", owner), zap.Any("projectId", id))
	err = storage.AcceptProject(id)
	if err != nil {
		logger.Log.Warn("Failed to update project in storage", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update project in storage [%s]", err), http.StatusBadGateway)
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
	id, err := storage.CreateBid(projectId, owner, bid.Price, bid.Deadline, bid.Message)
	if err != nil {
		logger.Log.Warn("Failed to create bid in storage", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to create bid in storage [%s]", err), http.StatusBadRequest)
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
	bid, err := storage.GetBid(bidId)
	if err != nil {
		logger.Log.Warn("invalid bid id")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	project, err := storage.GetProject(bid.Project)
	if err != nil {
		logger.Log.Warn("invalid project in bid")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if project.Owner != owner {
		logger.Log.Warn("invalid user", zap.String("user", owner),
			zap.String("projectOwner", project.Owner))
		http.Error(w, "invalid user", http.StatusForbidden)
		return

	}

	logger.Log.Info("acceptBid", zap.String("user", owner), zap.Any("bid", bid))
	err = storage.AcceptBid(bidId)
	if err != nil {
		logger.Log.Warn("Failed to accept bid in storage", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to accept bid in storage [%s]", err), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	project, err = storage.GetProject(bid.Project)
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

	savedBid, err := storage.GetBid(bid.Id)
	if err != nil {
		logger.Log.Warn("Failed to load saved bid", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to load saved bid [%s]", err), http.StatusBadRequest)
		return
	}

	if savedBid.User != owner {
		logger.Log.Warn("Invalid owner", zap.String("bid owner", savedBid.User),
			zap.String("request sender", owner))
		http.Error(w, "Invalid bid owner", http.StatusForbidden)
		return
	}

	logger.Log.Info("updateBid", zap.String("user", owner), zap.Any("bid", bid))
	err = storage.UpdateBid(bid.Id, bid.Price, bid.Deadline, bid.Message)
	if err != nil {
		logger.Log.Warn("Failed to update bid in storage", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to update bid in storage [%s]", err), http.StatusInternalServerError)
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
	owner := r.Header.Get("TRTG-Address")
	if len(owner) == 0 {
		logger.Log.Warn("no owner in headers")
		http.Error(w, "no public key in headers", http.StatusBadRequest)
		return
	}

	savedBid, err := storage.GetBid(bidId)
	if err != nil {
		logger.Log.Warn("Failed to load bid", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to load bid [%s]", err), http.StatusBadRequest)
		return
	}

	if savedBid.User != owner {
		logger.Log.Warn("Invalid owner", zap.String("bid owner", savedBid.User),
			zap.String("request sender", owner))
		http.Error(w, "Invalid bid owner", http.StatusForbidden)
		return
	}

	logger.Log.Info("deleteBid", zap.String("user", owner), zap.Any("bidId", bidId))
	err = storage.DeleteBid(bidId)
	if err != nil {
		logger.Log.Warn("Failed to delete bid in storage", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to delete bid in storage [%s]", err), http.StatusBadGateway)
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
