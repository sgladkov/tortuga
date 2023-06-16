package storage

import (
	"database/sql"
	"sync"

	"github.com/sgladkov/tortuga/internal/logger"
	"github.com/sgladkov/tortuga/internal/models"
	"go.uber.org/zap"
)

type PgStorage struct {
	lock sync.Mutex
	db   *sql.DB
}

func NewPgStorage(db *sql.DB) (*PgStorage, error) {
	err := initDB(db)
	if err != nil {
		logger.Log.Error("Failed to init database", zap.Error(err))
		return nil, err
	}
	return &PgStorage{
		db: db,
	}, nil
}

func initDB(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		logger.Log.Error("Failed to create db transaction", zap.Error(err))
		return err
	}
	defer func() {
		err = tx.Rollback()
		if err != nil {
			logger.Log.Info("Failed to rollback db transaction", zap.String("error", err.Error()))
		}
	}()

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS Users (" +
		"id varchar(42) PRIMARY KEY, " +
		"nickname varchar(256), " +
		"description text, " +
		"nonce bigint, " +
		"registered timestamp, " +
		"status smallint, " +
		"tags text, " +
		"rating double precision, " +
		"account bigint)")
	if err != nil {
		logger.Log.Error("Failed to create db table", zap.Error(err))
		return err
	}
	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS Projects (" +
		"id bigint PRIMARY KEY, " +
		"title varchar(1024), " +
		"description text, " +
		"tags text, " +
		"created timestamp, " +
		"status smallint, " +
		"owner varchar(42) REFERENCES Users(id), " +
		"contractor varchar(42) REFERENCES Users(id), " +
		"started timestamp, " +
		"deadline interval, " +
		"price bigint)")
	if err != nil {
		logger.Log.Error("Failed to create db table", zap.Error(err))
		return err
	}
	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS Bids (" +
		"project bigint REFERENCES Projects(id), " +
		"user varchar(42) REFERENCES Users(id), " +
		"price bigint, " +
		"deadline interval, " +
		"message text, " +
		"PRIMARY KEY(project, user))")
	if err != nil {
		logger.Log.Error("Failed to create db table", zap.Error(err))
		return err
	}
	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS Rates (" +
		"author varchar(42) REFERENCES Users(id), " +
		"project bigint REFERENCES Projects(id), " +
		"user varchar(42) REFERENCES Users(id), " +
		"rate smallint, " +
		"message text, " +
		"PRIMARY KEY(author, project))")
	if err != nil {
		logger.Log.Error("Failed to create db table", zap.Error(err))
		return err
	}

	return tx.Commit()
}

func (s *PgStorage) GetUserList() (*models.UserList, error) {
	res := models.UserList{}
	return &res, nil
}

func (s *PgStorage) GetUser(id string) (*models.User, error) {
	res := models.User{}
	return &res, nil
}

func (s *PgStorage) GetProjectList() (*models.ProjectList, error) {
	res := models.ProjectList{}
	return &res, nil
}

func (s *PgStorage) GetUserProjects(userId string) (*models.ProjectList, error) {
	res := models.ProjectList{}
	return &res, nil
}

func (s *PgStorage) GetProject(id uint64) (*models.Project, error) {
	res := models.Project{}
	return &res, nil
}

func (s *PgStorage) AddUser(user *models.User) error {
	return nil
}

func (s *PgStorage) UpdateUserNonce(id string, nonce uint64) error {
	return nil
}
