package storage

import (
	"github.com/sgladkov/tortuga/internal/models"
)

type Storage interface {
	BeginTx() error
	CommitTx() error
	RollbackTx() error

	CreateUser(user models.User) error
	GetUserList() ([]models.User, error)
	GetUser(id string) (models.User, error)
	UpdateUser(user models.User) error
	DeleteUser(id string) error

	CreateProject(project models.Project) (uint64, error)
	GetProjectList() ([]models.Project, error)
	GetUserProjects(userId string) ([]models.Project, error)
	GetProject(id uint64) (models.Project, error)
	UpdateProject(project models.Project) error
	DeleteProject(projectId uint64) error

	CreateBid(bid models.Bid) (uint64, error)
	GetBid(id uint64) (models.Bid, error)
	GetProjectBids(projectId uint64) ([]models.Bid, error)
	UpdateBid(bid models.Bid) error
	DeleteBid(id uint64) error

	Close() error
}
