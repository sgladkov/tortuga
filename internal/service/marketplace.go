package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/sgladkov/tortuga/internal/logger"
	"github.com/sgladkov/tortuga/internal/models"
	"github.com/sgladkov/tortuga/internal/storage"
	"go.uber.org/zap"
	"time"
)

type Marketplace struct {
	storage storage.Storage
}

func NewMarketplace(s storage.Storage) *Marketplace {
	return &Marketplace{
		storage: s,
	}
}

func (m *Marketplace) GetUserList(ctx context.Context) ([]models.User, error) {
	return m.storage.GetUserList(ctx)
}

func (m *Marketplace) GetUser(ctx context.Context, id string) (models.User, error) {
	return m.storage.GetUser(ctx, id)
}

func (m *Marketplace) GetProjectList(ctx context.Context) ([]models.Project, error) {
	return m.storage.GetProjectList(ctx)
}

func (m *Marketplace) GetUserProjects(ctx context.Context, userId string) ([]models.Project, error) {
	return m.storage.GetUserProjects(ctx, userId)
}

func (m *Marketplace) GetProject(ctx context.Context, id uint64) (models.Project, error) {
	return m.storage.GetProject(ctx, id)
}

func (m *Marketplace) AddUser(ctx context.Context, caller string, user models.User) error {
	if caller != user.Id {
		logger.Log.Warn("forbidden", zap.String("caller", caller), zap.String("required", user.Id))
		return fmt.Errorf("forbidden for caller %s", caller)
	}
	return m.storage.CreateUser(ctx, user)
}

func (m *Marketplace) UpdateUserNonce(ctx context.Context, id string, nonce uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	user, err := m.storage.GetUser(ctx, id)
	if err != nil {
		return err
	}
	user.Nonce = nonce
	err = m.storage.UpdateUser(ctx, user)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}

func (m *Marketplace) CreateProject(ctx context.Context, caller string, title string, description string, tags models.Tags, owner string, deadline time.Duration,
	price uint64) (uint64, error) {
	if caller != owner {
		logger.Log.Warn("forbidden", zap.String("caller", caller), zap.String("required", owner))
		return 0, fmt.Errorf("forbidden for caller %s", caller)
	}
	p := models.Project{
		Title:       title,
		Description: description,
		Tags:        tags,
		Created:     time.Now(),
		Status:      models.Open,
		Owner:       owner,
		Deadline:    deadline,
		Price:       price,
	}
	return m.storage.CreateProject(ctx, p)
}

func (m *Marketplace) UpdateProject(ctx context.Context, caller string, projectId uint64, title string, description string, tags models.Tags, deadline time.Duration,
	price uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	project, err := m.storage.GetProject(ctx, projectId)
	if err != nil {
		return err
	}

	if caller != project.Owner {
		logger.Log.Warn("forbidden", zap.String("caller", caller), zap.String("required", project.Owner))
		return fmt.Errorf("forbidden for caller %s", caller)
	}
	if project.Status != models.Open {
		logger.Log.Warn("invalid project status", zap.Uint64("id", projectId),
			zap.Uint8("status", uint8(project.Status)))
		return errors.New("forbidden for project in current status")
	}

	project.Title = title
	project.Description = description
	project.Tags = tags
	project.Deadline = deadline
	project.Price = price
	err = m.storage.UpdateProject(ctx, project)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}

func (m *Marketplace) DeleteProject(ctx context.Context, caller string, projectId uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	project, err := m.storage.GetProject(ctx, projectId)
	if err != nil {
		return err
	}

	if caller != project.Owner {
		logger.Log.Warn("forbidden", zap.String("caller", caller), zap.String("required", project.Owner))
		return fmt.Errorf("forbidden for caller %s", caller)
	}
	if project.Status != models.Open {
		logger.Log.Warn("invalid project status", zap.Uint64("id", projectId),
			zap.Uint8("status", uint8(project.Status)))
		return errors.New("forbidden for project in current status")
	}

	err = m.storage.DeleteProject(ctx, projectId)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}

func (m *Marketplace) CreateBid(ctx context.Context, caller string, projectId uint64, fromUser string, price uint64, deadline time.Duration,
	message string) (uint64, error) {
	err := m.storage.BeginTx()
	if err != nil {
		return 0, err
	}
	defer m.storage.RollbackTx()

	project, err := m.storage.GetProject(ctx, projectId)
	if err != nil {
		return 0, fmt.Errorf("no project %v", projectId)
	}
	_, err = m.storage.GetUser(ctx, fromUser)
	if err != nil {
		return 0, fmt.Errorf("no user %v", fromUser)
	}

	if caller != fromUser {
		logger.Log.Warn("forbidden", zap.String("caller", caller), zap.String("required", fromUser))
		return 0, fmt.Errorf("forbidden for caller %s", caller)
	}
	if caller == project.Owner {
		logger.Log.Warn("bidder is project owner", zap.String("caller", caller))
		return 0, fmt.Errorf("forbidden for caller %s", caller)
	}
	if project.Status != models.Open {
		logger.Log.Warn("invalid project status", zap.Uint64("id", projectId),
			zap.Uint8("status", uint8(project.Status)))
		return 0, errors.New("forbidden for project in current status")
	}

	bid := models.Bid{
		Project:  projectId,
		User:     fromUser,
		Deadline: deadline,
		Price:    price,
		Message:  message,
	}
	bidId, err := m.storage.CreateBid(ctx, bid)
	if err != nil {
		return 0, err
	}

	err = m.storage.CommitTx()
	if err != nil {
		return 0, err
	}
	return bidId, nil
}

func (m *Marketplace) GetBid(ctx context.Context, id uint64) (models.Bid, error) {
	return m.storage.GetBid(ctx, id)
}

func (m *Marketplace) GetProjectBids(ctx context.Context, projectId uint64) ([]models.Bid, error) {
	return m.storage.GetProjectBids(ctx, projectId)
}

func (m *Marketplace) UpdateBid(ctx context.Context, caller string, id uint64, price uint64, deadline time.Duration, message string) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	bid, err := m.storage.GetBid(ctx, id)
	if err != nil {
		return err
	}

	if caller != bid.User {
		logger.Log.Warn("forbidden", zap.String("caller", caller), zap.String("required", bid.User))
		return fmt.Errorf("forbidden for caller %s", caller)
	}

	bid.Price = price
	bid.Deadline = deadline
	bid.Message = message
	err = m.storage.UpdateBid(ctx, bid)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}

func (m *Marketplace) DeleteBid(ctx context.Context, caller string, id uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	bid, err := m.storage.GetBid(ctx, id)
	if err != nil {
		return err
	}

	if caller != bid.User {
		logger.Log.Warn("forbidden", zap.String("caller", caller), zap.String("required", bid.User))
		return fmt.Errorf("forbidden for caller %s", caller)
	}

	err = m.storage.DeleteBid(ctx, id)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}

func (m *Marketplace) AcceptBid(ctx context.Context, caller string, id uint64) (uint64, error) {
	err := m.storage.BeginTx()
	if err != nil {
		return 0, err
	}
	defer m.storage.RollbackTx()

	bid, err := m.storage.GetBid(ctx, id)
	if err != nil {
		return 0, err
	}
	project, err := m.storage.GetProject(ctx, bid.Project)
	if err != nil {
		return 0, err
	}

	if caller != bid.User {
		logger.Log.Warn("forbidden", zap.String("caller", caller), zap.String("required", bid.User))
		return 0, fmt.Errorf("forbidden for caller %s", caller)
	}
	if project.Status != models.Open {
		logger.Log.Warn("invalid project status", zap.Uint64("id", bid.Project),
			zap.Uint8("status", uint8(project.Status)))
		return 0, errors.New("forbidden for project in current status")
	}

	project.Contractor = bid.User
	project.Started = time.Now()
	project.Deadline = bid.Deadline
	project.Price = bid.Price
	project.Status = models.InWork
	err = m.storage.UpdateProject(ctx, project)
	if err != nil {
		return 0, err
	}
	err = m.storage.DeleteBid(ctx, id)
	if err != nil {
		return 0, err
	}

	return bid.Project, m.storage.CommitTx()
}

func (m *Marketplace) CancelProject(ctx context.Context, caller string, id uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	project, err := m.storage.GetProject(ctx, id)
	if err != nil {
		return err
	}

	if caller != project.Owner && caller != project.Contractor {
		logger.Log.Warn("forbidden", zap.String("caller", caller),
			zap.String("owner", project.Owner), zap.String("contractor", project.Contractor))
		return fmt.Errorf("forbidden for caller %s", caller)
	}
	if project.Status == models.InWork || project.Status == models.InReview {
		project.Status = models.Canceled
	} else {
		return fmt.Errorf("invalid project status %d", project.Status)
	}

	err = m.storage.UpdateProject(ctx, project)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}

func (m *Marketplace) SetProjectReady(ctx context.Context, caller string, id uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	project, err := m.storage.GetProject(ctx, id)
	if err != nil {
		return err
	}

	if caller != project.Contractor {
		logger.Log.Warn("forbidden", zap.String("caller", caller),
			zap.String("required", project.Contractor))
		return fmt.Errorf("forbidden for caller %s", caller)
	}
	if project.Status == models.InWork {
		project.Status = models.InReview
	} else {
		return fmt.Errorf("invalid project status %d", project.Status)
	}

	err = m.storage.UpdateProject(ctx, project)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}

func (m *Marketplace) AcceptProject(ctx context.Context, caller string, id uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	project, err := m.storage.GetProject(ctx, id)
	if err != nil {
		return err
	}

	if caller != project.Owner {
		logger.Log.Warn("forbidden", zap.String("caller", caller), zap.String("required", project.Owner))
		return fmt.Errorf("forbidden for caller %s", caller)
	}
	if project.Status == models.InReview {
		project.Status = models.Completed
	} else {
		return fmt.Errorf("invalid project status %d", project.Status)
	}

	err = m.storage.UpdateProject(ctx, project)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}
