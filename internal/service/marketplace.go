package service

import (
	"context"
	"fmt"
	"github.com/sgladkov/tortuga/internal/models"
	"github.com/sgladkov/tortuga/internal/storage"
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

func (m *Marketplace) AddUser(ctx context.Context, user models.User) error {
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

func (m *Marketplace) CreateProject(ctx context.Context, title string, description string, tags models.Tags, owner string, deadline time.Duration,
	price uint64) (uint64, error) {
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

func (m *Marketplace) UpdateProject(ctx context.Context, projectId uint64, title string, description string, tags models.Tags, deadline time.Duration,
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

func (m *Marketplace) DeleteProject(ctx context.Context, projectId uint64) error {
	return m.storage.DeleteProject(ctx, projectId)
}

func (m *Marketplace) CreateBid(ctx context.Context, projectId uint64, fromUser string, price uint64, deadline time.Duration,
	message string) (uint64, error) {
	err := m.storage.BeginTx()
	if err != nil {
		return 0, err
	}
	defer m.storage.RollbackTx()

	_, err = m.storage.GetProject(ctx, projectId)
	if err != nil {
		return 0, fmt.Errorf("no project %v", projectId)
	}
	_, err = m.storage.GetUser(ctx, fromUser)
	if err != nil {
		return 0, fmt.Errorf("no user %v", fromUser)
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

func (m *Marketplace) UpdateBid(ctx context.Context, id uint64, price uint64, deadline time.Duration, message string) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	bid, err := m.storage.GetBid(ctx, id)
	if err != nil {
		return err
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

func (m *Marketplace) DeleteBid(ctx context.Context, id uint64) error {
	return m.storage.DeleteBid(ctx, id)
}

func (m *Marketplace) AcceptBid(ctx context.Context, id uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	bid, err := m.storage.GetBid(ctx, id)
	if err != nil {
		return err
	}
	project, err := m.storage.GetProject(ctx, bid.Project)
	if err != nil {
		return err
	}
	project.Contractor = bid.User
	project.Started = time.Now()
	project.Deadline = bid.Deadline
	project.Price = bid.Price
	project.Status = models.InWork
	err = m.storage.UpdateProject(ctx, project)
	if err != nil {
		return err
	}
	err = m.storage.DeleteBid(ctx, id)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}

func (m *Marketplace) CancelProject(ctx context.Context, id uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	project, err := m.storage.GetProject(ctx, id)
	if err != nil {
		return err
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

func (m *Marketplace) SetProjectReady(ctx context.Context, id uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	project, err := m.storage.GetProject(ctx, id)
	if err != nil {
		return err
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

func (m *Marketplace) AcceptProject(ctx context.Context, id uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	project, err := m.storage.GetProject(ctx, id)
	if err != nil {
		return err
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
