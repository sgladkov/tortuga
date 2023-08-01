package service

import (
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

func (m *Marketplace) GetUserList() ([]models.User, error) {
	return m.storage.GetUserList()
}

func (m *Marketplace) GetUser(id string) (models.User, error) {
	return m.storage.GetUser(id)
}

func (m *Marketplace) GetProjectList() ([]models.Project, error) {
	return m.storage.GetProjectList()
}

func (m *Marketplace) GetUserProjects(userId string) ([]models.Project, error) {
	return m.storage.GetUserProjects(userId)
}

func (m *Marketplace) GetProject(id uint64) (models.Project, error) {
	return m.storage.GetProject(id)
}

func (m *Marketplace) AddUser(user models.User) error {
	return m.storage.CreateUser(user)
}

func (m *Marketplace) UpdateUserNonce(id string, nonce uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	user, err := m.storage.GetUser(id)
	if err != nil {
		return err
	}
	user.Nonce = nonce
	err = m.storage.UpdateUser(user)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}

func (m *Marketplace) CreateProject(title string, description string, tags models.Tags, owner string, deadline time.Duration,
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
	return m.storage.CreateProject(p)
}

func (m *Marketplace) UpdateProject(projectId uint64, title string, description string, tags models.Tags, deadline time.Duration,
	price uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	project, err := m.storage.GetProject(projectId)
	if err != nil {
		return err
	}
	project.Title = title
	project.Description = description
	project.Tags = tags
	project.Deadline = deadline
	project.Price = price
	err = m.storage.UpdateProject(project)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}

func (m *Marketplace) DeleteProject(projectId uint64) error {
	return m.storage.DeleteProject(projectId)
}

func (m *Marketplace) CreateBid(projectId uint64, fromUser string, price uint64, deadline time.Duration,
	message string) (uint64, error) {
	err := m.storage.BeginTx()
	if err != nil {
		return 0, err
	}
	defer m.storage.RollbackTx()

	_, err = m.storage.GetProject(projectId)
	if err != nil {
		return 0, fmt.Errorf("no project %v", projectId)
	}
	_, err = m.storage.GetUser(fromUser)
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
	bidId, err := m.storage.CreateBid(bid)
	if err != nil {
		return 0, err
	}

	err = m.storage.CommitTx()
	if err != nil {
		return 0, err
	}
	return bidId, nil
}

func (m *Marketplace) GetBid(id uint64) (models.Bid, error) {
	return m.storage.GetBid(id)
}

func (m *Marketplace) GetProjectBids(projectId uint64) ([]models.Bid, error) {
	return m.storage.GetProjectBids(projectId)
}

func (m *Marketplace) UpdateBid(id uint64, price uint64, deadline time.Duration, message string) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	bid, err := m.storage.GetBid(id)
	if err != nil {
		return err
	}
	bid.Price = price
	bid.Deadline = deadline
	bid.Message = message
	err = m.storage.UpdateBid(bid)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}

func (m *Marketplace) DeleteBid(id uint64) error {
	return m.storage.DeleteBid(id)
}

func (m *Marketplace) AcceptBid(id uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	bid, err := m.storage.GetBid(id)
	if err != nil {
		return err
	}
	project, err := m.storage.GetProject(bid.Project)
	if err != nil {
		return err
	}
	project.Contractor = bid.User
	project.Started = time.Now()
	project.Deadline = bid.Deadline
	project.Price = bid.Price
	project.Status = models.InWork
	err = m.storage.UpdateProject(project)
	if err != nil {
		return err
	}
	err = m.storage.DeleteBid(id)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}

func (m *Marketplace) CancelProject(id uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	project, err := m.storage.GetProject(id)
	if err != nil {
		return err
	}
	if project.Status == models.InWork || project.Status == models.InReview {
		project.Status = models.Canceled
	} else {
		return fmt.Errorf("invalid project status %d", project.Status)
	}

	err = m.storage.UpdateProject(project)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}

func (m *Marketplace) SetProjectReady(id uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	project, err := m.storage.GetProject(id)
	if err != nil {
		return err
	}
	if project.Status == models.InWork {
		project.Status = models.InReview
	} else {
		return fmt.Errorf("invalid project status %d", project.Status)
	}

	err = m.storage.UpdateProject(project)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}

func (m *Marketplace) AcceptProject(id uint64) error {
	err := m.storage.BeginTx()
	if err != nil {
		return err
	}
	defer m.storage.RollbackTx()

	project, err := m.storage.GetProject(id)
	if err != nil {
		return err
	}
	if project.Status == models.InReview {
		project.Status = models.Completed
	} else {
		return fmt.Errorf("invalid project status %d", project.Status)
	}

	err = m.storage.UpdateProject(project)
	if err != nil {
		return err
	}

	return m.storage.CommitTx()
}
