package storage

import (
	"github.com/sgladkov/tortuga/internal/models"
	"time"
)

type Storage interface {
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
	Close() error
}
