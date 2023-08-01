package storage

import (
	"database/sql"
	"fmt"
	"github.com/sgladkov/tortuga/internal/logger"
	"github.com/sgladkov/tortuga/internal/models"
	"go.uber.org/zap"
)

type DBTX interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Prepare(string) (*sql.Stmt, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
}

type PgStorage struct {
	db   *sql.DB
	exec DBTX
	tx   *sql.Tx
}

func NewPgStorage(db *sql.DB) (*PgStorage, error) {
	err := initDB(db)
	if err != nil {
		logger.Log.Error("Failed to init database", zap.Error(err))
		return nil, err
	}
	return &PgStorage{
		db:   db,
		exec: db,
	}, nil
}

func initDB(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		logger.Log.Error("Failed to create exec transaction", zap.Error(err))
		return err
	}
	defer func() {
		err = tx.Rollback()
		if err != nil {
			logger.Log.Info("Failed to rollback exec transaction", zap.String("error", err.Error()))
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
		logger.Log.Error("Failed to create exec table", zap.Error(err))
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
		logger.Log.Error("Failed to create exec table", zap.Error(err))
		return err
	}
	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS Bids (" +
		"id bigint PRIMARY KEY generated always as identity, " +
		"project bigint REFERENCES Projects(id), " +
		"user varchar(42) REFERENCES Users(id), " +
		"price bigint, " +
		"deadline interval, " +
		"message text)")
	if err != nil {
		logger.Log.Error("Failed to create exec table", zap.Error(err))
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
		logger.Log.Error("Failed to create exec table", zap.Error(err))
		return err
	}

	return tx.Commit()
}

func (s *PgStorage) BeginTx() error {
	if s.tx != nil {
		logger.Log.Error("attempt to create embedded transaction")
		return fmt.Errorf("attempt to create embedded transaction")
	}
	tx, err := s.db.Begin()
	if err != nil {
		logger.Log.Error("Failed to create exec transaction", zap.Error(err))
		return err
	}
	s.exec = tx
	s.tx = tx
	return nil
}

func (s *PgStorage) CommitTx() error {
	if s.tx == nil {
		logger.Log.Error("attempt to commit closed transaction")
		return fmt.Errorf("attempt to commit closed transaction")
	}
	err := s.tx.Commit()
	if err != nil {
		logger.Log.Error("Failed to commit exec transaction", zap.Error(err))
		return err
	}
	s.exec = s.db
	s.tx = nil
	return nil
}

func (s *PgStorage) RollbackTx() error {
	if s.tx == nil {
		logger.Log.Error("attempt to rollback closed transaction")
		return fmt.Errorf("attempt to rollback closed transaction")
	}
	err := s.tx.Rollback()
	if err != nil {
		logger.Log.Error("Failed to rollback exec transaction", zap.Error(err))
		return err
	}
	s.exec = s.db
	s.tx = nil
	return nil
}

func (s *PgStorage) GetUserList() ([]models.User, error) {
	rowsUsers, err := s.exec.Query("SELECT id, nickname, description, nonce, registered, status, tags, " +
		"rating, account FROM Users")
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
	var res []models.User
	for rowsUsers.Next() {
		var user models.User
		err = rowsUsers.Scan(&user.Id, &user.Nickname, &user.Description, &user.Nonce, &user.Registered,
			&user.Status, &user.Tags, &user.Rating, &user.Account)
		if err != nil {
			logger.Log.Error("Failed to get data from rowset", zap.Error(err))
			return nil, err
		}
		res = append(res, user)
	}
	err = rowsUsers.Err()
	if err != nil {
		logger.Log.Error("error while iterating rows", zap.Error(err))
		return nil, err
	}
	return res, nil
}

func (s *PgStorage) GetUser(id string) (models.User, error) {
	// TODO: set explicit field set and order in query
	stmtUser, err := s.exec.Prepare("SELECT * FROM Users WHERE id = $1")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return models.User{}, err
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
		return models.User{}, err
	}
	res := models.User{}
	err = rowUser.Scan(&res.Id, &res.Nickname, &res.Description, &res.Nonce, &res.Registered, &res.Status,
		&res.Tags, &res.Rating, &res.Account)
	if err != nil {
		logger.Log.Error("Failed to get data from rowset", zap.Error(err))
		return models.User{}, err
	}
	return res, nil
}

func (s *PgStorage) GetProjectList() ([]models.Project, error) {
	rowsProjects, err := s.exec.Query("SELECT id, title, description, tags, created, status, owner, " +
		"contractor, started, deadline, price FROM Projects")
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
	var res []models.Project
	for rowsProjects.Next() {
		var project models.Project
		err = rowsProjects.Scan(&project.Id, &project.Title, &project.Description, &project.Tags,
			&project.Created, &project.Status, &project.Owner, &project.Contractor, &project.Started,
			&project.Deadline, &project.Price)
		if err != nil {
			logger.Log.Error("Failed to get data from rowset", zap.Error(err))
			return nil, err
		}
		res = append(res, project)
	}
	err = rowsProjects.Err()
	if err != nil {
		logger.Log.Error("error while iterating rows", zap.Error(err))
		return nil, err
	}
	return res, nil
}

func (s *PgStorage) GetUserProjects(userId string) ([]models.Project, error) {
	stmtProjects, err := s.exec.Prepare("SELECT id, title, description, tags, created, status, owner, " +
		"contractor, started, deadline, price  FROM Projects WHERE owner = $1")
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
	var res []models.Project
	for rowsProjects.Next() {
		var project models.Project
		err = rowsProjects.Scan(&project.Id, &project.Title, &project.Description, &project.Tags,
			&project.Created, &project.Status, &project.Owner, &project.Contractor, &project.Started,
			&project.Deadline, &project.Price)
		if err != nil {
			logger.Log.Error("Failed to get data from rowset", zap.Error(err))
			return nil, err
		}
		res = append(res, project)
	}
	err = rowsProjects.Err()
	if err != nil {
		logger.Log.Error("error while iterating rows", zap.Error(err))
		return nil, err
	}
	return res, nil
}

func (s *PgStorage) GetProject(id uint64) (models.Project, error) {
	// TODO: set explicit field set and order in query
	stmtProject, err := s.exec.Prepare("SELECT * FROM Projects WHERE id = $1")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return models.Project{}, err
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
		return models.Project{}, err
	}
	res := models.Project{}
	err = rowProject.Scan(&res.Id, &res.Title, &res.Description, &res.Tags, &res.Created, &res.Status,
		&res.Owner, &res.Contractor, &res.Started, &res.Deadline, &res.Price)
	if err != nil {
		logger.Log.Error("Failed to get data from rowset", zap.Error(err))
		return models.Project{}, err
	}
	return res, nil
}

func (s *PgStorage) CreateUser(user models.User) error {
	stmtUser, err := s.exec.Prepare("INSERT INTO Users (id, nickname, description, nonce, registered, " +
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

func (s *PgStorage) UpdateUser(user models.User) error {
	stmtUser, err := s.exec.Prepare("UPDATE Users SET nickname=$1, description=$2, nonce=$3, registered=$4, " +
		"status=$5, tags=$6, rating=$7, account=$8 WHERE id=$9")
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

	_, err = stmtUser.Exec(user.Nickname, user.Description, user.Nonce, user.Registered, user.Status,
		user.Tags, user.Rating, user.Account, user.Id)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}

func (s *PgStorage) CreateProject(project models.Project) (uint64, error) {
	stmtUser, err := s.exec.Prepare("INSERT INTO Projects (title, description, tags, created, status, owner, deadline, price) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id")
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

	row := stmtUser.QueryRow(project.Title, project.Description, project.Tags, project.Created, project.Status, project.Owner, project.Deadline, project.Price)
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

func (s *PgStorage) UpdateProject(project models.Project) error {
	stmtUser, err := s.exec.Prepare("UPDATE Projects SET title=$1, description=$2, tags=$3, created=$4, " +
		"status=$5, owner=$6, contractor=$7, started=$8, deadline=$9, price=$10 WHERE id=$11")
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

	_, err = stmtUser.Exec(project.Title, project.Description, project.Tags, project.Created, project.Status,
		project.Owner, project.Contractor, project.Started, project.Deadline, project.Price, project.Id)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}

func (s *PgStorage) DeleteUser(id string) error {
	stmtUser, err := s.exec.Prepare("DELETE FROM Users WHERE id = $1")
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

	_, err = stmtUser.Exec(id)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}

func (s *PgStorage) DeleteProject(projectId uint64) error {
	stmtUser, err := s.exec.Prepare("DELETE FROM Projects WHERE id = $1")
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

	_, err = stmtUser.Exec(projectId)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}

func (s *PgStorage) CreateBid(bid models.Bid) (uint64, error) {
	stmtBid, err := s.exec.Prepare("INSERT INTO Bids (project, user, price, deadline, message) VALUES($1, $2, $3, $4, $5) RETURNING id")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return 0, err
	}
	defer func() {
		err = stmtBid.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()

	row := stmtBid.QueryRow(bid.Id, bid.User, bid.Price, bid.Deadline, bid.Message)
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

func (s *PgStorage) GetBid(id uint64) (models.Bid, error) {
	// TODO: set explicit field set and order in query
	stmtBid, err := s.exec.Prepare("SELECT * FROM Bids WHERE id = $1")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return models.Bid{}, err
	}
	defer func() {
		err = stmtBid.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()
	rowBid := stmtBid.QueryRow(id)
	if err = rowBid.Err(); err != nil {
		logger.Log.Error("Failed to query project data", zap.Error(err))
		return models.Bid{}, err
	}
	res := models.Bid{}
	err = rowBid.Scan(&res.Id, &res.Project, &res.User, &res.Price, &res.Deadline, &res.Message)
	if err != nil {
		logger.Log.Error("Failed to get data from rowset", zap.Error(err))
		return models.Bid{}, err
	}
	return res, nil
}

func (s *PgStorage) GetProjectBids(projectId uint64) ([]models.Bid, error) {
	stmtBids, err := s.exec.Prepare("SELECT id, project, user, price, deadline, message FROM Bids WHERE project = $1")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return nil, err
	}
	defer func() {
		err = stmtBids.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()
	rowsBids, err := stmtBids.Query(projectId)
	if err != nil {
		logger.Log.Error("Failed to query Users data", zap.Error(err))
		return nil, err
	}
	defer func() {
		err = rowsBids.Close()
		if err != nil {
			logger.Log.Error("Failed to close rowset", zap.Error(err))
		}
	}()
	var res []models.Bid
	for rowsBids.Next() {
		var bid models.Bid
		err = rowsBids.Scan(&bid.Id, &bid.Project, &bid.User, &bid.Price, &bid.Deadline, &bid.Message)
		if err != nil {
			logger.Log.Error("Failed to get data from rowset", zap.Error(err))
			return nil, err
		}
		res = append(res, bid)
	}
	err = rowsBids.Err()
	if err != nil {
		logger.Log.Error("error while iterating rows", zap.Error(err))
		return nil, err
	}
	return res, nil
}

func (s *PgStorage) UpdateBid(bid models.Bid) error {
	stmtBid, err := s.exec.Prepare("UPDATE Bids SET project=$1, user=$2, price=$3, deadline = $4, message = $5 WHERE id = $6")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return err
	}
	defer func() {
		err = stmtBid.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()

	_, err = stmtBid.Exec(bid.Project, bid.User, bid.Price, bid.Deadline, bid.Message, bid.Id)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}

func (s *PgStorage) DeleteBid(id uint64) error {
	stmtBid, err := s.exec.Prepare("DELETE FROM Bids WHERE id = $1")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return err
	}
	defer func() {
		err = stmtBid.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()

	_, err = stmtBid.Exec(id)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}

func (s *PgStorage) Close() error {
	return s.db.Close()
}
