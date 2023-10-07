package storage

import (
	"context"
	"github.com/sgladkov/tortuga/internal/models"
)

type Storage interface {
	BeginTx() error
	CommitTx() error
	RollbackTx() error

	CreateUser(ctx context.Context, user models.User) error
	GetUserList(ctx context.Context) ([]models.User, error)
	GetUser(ctx context.Context, id string) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) error
	DeleteUser(ctx context.Context, id string) error

	CreateProject(ctx context.Context, project models.Project) (uint64, error)
	GetProjectList(ctx context.Context) ([]models.Project, error)
	GetUserProjects(ctx context.Context, userId string) ([]models.Project, error)
	GetProject(ctx context.Context, id uint64) (models.Project, error)
	UpdateProject(ctx context.Context, project models.Project) error
	DeleteProject(ctx context.Context, projectId uint64) error

	CreateBid(ctx context.Context, bid models.Bid) (uint64, error)
	GetBid(ctx context.Context, id uint64) (models.Bid, error)
	GetProjectBids(ctx context.Context, projectId uint64) ([]models.Bid, error)
	UpdateBid(ctx context.Context, bid models.Bid) error
	DeleteBid(ctx context.Context, id uint64) error

	CreateRate(ctx context.Context, rate models.Rate) (uint64, error)
	GetRate(ctx context.Context, id uint64) (models.Rate, error)
	GetEvaluatorRates(ctx context.Context, userId string) ([]models.Rate, error)
	GetEvaluatedRates(ctx context.Context, userId string) ([]models.Rate, error)
	DeleteRate(ctx context.Context, id uint64) error

	Close() error
}
