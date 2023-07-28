package storage

import (
	"github.com/sgladkov/tortuga/internal/models"
	"time"
)

type Storage interface {
	BeginTx() error
	CommitTx() error
	RollbackTx() error
	GetUserList() (*models.UserList, error)
	GetUser(id string) (*models.User, error)
	GetProjectList() (*models.ProjectList, error)
	GetUserProjects(userId string) (*models.ProjectList, error)
	GetProject(id uint64) (*models.Project, error)
	AddUser(user *models.User) error
	UpdateUserNonce(id string, nonce uint64) error
	CreateProject(title string, description string, tags models.Tags, owner string, deadline time.Duration,
		price uint64) (uint64, error)
	UpdateProject(projectId uint64, title string, description string, tags models.Tags, deadline time.Duration,
		price uint64) error
	DeleteProject(projectId uint64) error
	CreateBid(projectId uint64, fromUser string, price uint64, deadline time.Duration, message string) (uint64, error)
	GetBid(id uint64) (*models.Bid, error)
	GetProjectBids(projectId uint64) ([]uint64, error)
	UpdateBid(id uint64, price uint64, deadline time.Duration, message string) error
	DeleteBid(id uint64) error
	AcceptBid(id uint64) error
	CancelProject(id uint64) error
	SetProjectReady(id uint64) error
	AcceptProject(id uint64) error
	Close() error
}
