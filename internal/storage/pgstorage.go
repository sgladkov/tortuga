package storage

import (
	"database/sql"
	"github.com/sgladkov/tortuga/internal/logger"
	"github.com/sgladkov/tortuga/internal/models"
	"time"

	"go.uber.org/zap"
)

type PgStorage struct {
	db *sql.DB
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
		"id bigint PRIMARY KEY generated always as identity, " +
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
	rowsUsers, err := s.db.Query("SELECT id FROM Users")
	if err != nil {
		logger.Log.Error("Failed to query Users data", zap.Error(err))
		return nil, err
	}
	defer func() {
		err = rowsUsers.Close()
		if err != nil {
			logger.Log.Error("Failed to close rowset", zap.Error(err))
		}
	}()
	res := models.UserList{}
	for rowsUsers.Next() {
		var id string
		err = rowsUsers.Scan(&id)
		if err != nil {
			logger.Log.Error("Failed to get data from rowset", zap.Error(err))
			return nil, err
		}
		res = append(res, id)
	}
	err = rowsUsers.Err()
	if err != nil {
		logger.Log.Error("error while iterating rows", zap.Error(err))
		return nil, err
	}
	return &res, nil
}

func (s *PgStorage) GetUser(id string) (*models.User, error) {
	// TODO: set explicit field set and order in query
	stmtUser, err := s.db.Prepare("SELECT * FROM Users WHERE id = $1")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return nil, err
	}
	defer func() {
		err = stmtUser.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()
	rowUser := stmtUser.QueryRow(id)
	if rowUser.Err() != nil {
		logger.Log.Error("Failed to query project data", zap.Error(err))
		return nil, err
	}
	res := models.User{}
	err = rowUser.Scan(&res.Id, &res.Nickname, &res.Description, &res.Nonce, &res.Registered, &res.Status,
		&res.Tags, &res.Rating, &res.Account)
	if err != nil {
		logger.Log.Error("Failed to get data from rowset", zap.Error(err))
		return nil, err
	}
	return &res, nil
}

func (s *PgStorage) GetProjectList() (*models.ProjectList, error) {
	rowsProjects, err := s.db.Query("SELECT id FROM Projects")
	if err != nil {
		logger.Log.Error("Failed to query Users data", zap.Error(err))
		return nil, err
	}
	defer func() {
		err = rowsProjects.Close()
		if err != nil {
			logger.Log.Error("Failed to close rowset", zap.Error(err))
		}
	}()
	res := models.ProjectList{}
	for rowsProjects.Next() {
		var id uint64
		err = rowsProjects.Scan(&id)
		if err != nil {
			logger.Log.Error("Failed to get data from rowset", zap.Error(err))
			return nil, err
		}
		res = append(res, id)
	}
	err = rowsProjects.Err()
	if err != nil {
		logger.Log.Error("error while iterating rows", zap.Error(err))
		return nil, err
	}
	return &res, nil
}

func (s *PgStorage) GetUserProjects(userId string) (*models.ProjectList, error) {
	stmtProjects, err := s.db.Prepare("SELECT id FROM Projects WHERE owner = $1")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return nil, err
	}
	defer func() {
		err = stmtProjects.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()
	rowsProjects, err := stmtProjects.Query(userId)
	if err != nil {
		logger.Log.Error("Failed to query Users data", zap.Error(err))
		return nil, err
	}
	defer func() {
		err = rowsProjects.Close()
		if err != nil {
			logger.Log.Error("Failed to close rowset", zap.Error(err))
		}
	}()
	res := models.ProjectList{}
	for rowsProjects.Next() {
		var id uint64
		err = rowsProjects.Scan(&id)
		if err != nil {
			logger.Log.Error("Failed to get data from rowset", zap.Error(err))
			return nil, err
		}
		res = append(res, id)
	}
	err = rowsProjects.Err()
	if err != nil {
		logger.Log.Error("error while iterating rows", zap.Error(err))
		return nil, err
	}
	return &res, nil
}

func (s *PgStorage) GetProject(id uint64) (*models.Project, error) {
	// TODO: set explicit field set and order in query
	stmtProject, err := s.db.Prepare("SELECT * FROM Projects WHERE id = $1")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return nil, err
	}
	defer func() {
		err = stmtProject.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()
	rowProject := stmtProject.QueryRow(id)
	if err = rowProject.Err(); err != nil {
		logger.Log.Error("Failed to query project data", zap.Error(err))
		return nil, err
	}
	res := models.Project{}
	err = rowProject.Scan(&res.Id, &res.Title, &res.Description, &res.Tags, &res.Created, &res.Status,
		&res.Owner, &res.Contractor, &res.Started, &res.Deadline, &res.Price)
	if err != nil {
		logger.Log.Error("Failed to get data from rowset", zap.Error(err))
		return nil, err
	}
	return &res, nil
}

func (s *PgStorage) AddUser(user *models.User) error {
	stmtUser, err := s.db.Prepare("INSERT INTO Users (id, nickname, description, nonce, registered, " +
		"status, tags, rating, account) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return err
	}
	defer func() {
		err = stmtUser.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()

	_, err = stmtUser.Exec(user.Id, user.Nickname, user.Description, user.Nonce, user.Registered,
		user.Status, user.Tags, user.Rating, user.Account)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}

func (s *PgStorage) UpdateUserNonce(id string, nonce uint64) error {
	stmtUser, err := s.db.Prepare("UPDATE Users SET nonce = $1 WHERE id=$2")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return err
	}
	defer func() {
		err = stmtUser.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()

	_, err = stmtUser.Exec(nonce, id)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}

func (s *PgStorage) CreateProject(title string, description string, tags models.Tags, owner string, deadline time.Duration, price uint64) (uint64, error) {
	stmtUser, err := s.db.Prepare("INSERT INTO Projects (title, description, tags, created, status, owner, deadline, price) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return 0, err
	}
	defer func() {
		err = stmtUser.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()

	row := stmtUser.QueryRow(title, description, tags, time.Now(), models.Open, owner, deadline, price)
	if err = row.Err(); err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return 0, err
	}

	var id uint64
	err = row.Scan(&id)
	if err != nil {
		logger.Log.Error("Failed to get data from rowset", zap.Error(err))
		return 0, err
	}
	return id, nil
}

func (s *PgStorage) Close() error {
	return s.db.Close()
}
