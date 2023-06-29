package storage

import "github.com/sgladkov/tortuga/internal/models"

type Storage interface {
	GetUserList() (*models.UserList, error)
	GetUser(id string) (*models.User, error)
	GetProjectList() (*models.ProjectList, error)
	GetUserProjects(userId string) (*models.ProjectList, error)
	GetProject(id uint64) (*models.Project, error)
	AddUser(user *models.User) error
	UpdateUserNonce(id string, nonce uint64) error
	Close() error
}
