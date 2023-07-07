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
		"id bigint PRIMARY KEY generated always as identity, " +
		"project bigint REFERENCES Projects(id), " +
		"user varchar(42) REFERENCES Users(id), " +
		"price bigint, " +
		"deadline interval, " +
		"message text)")
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

func (s *PgStorage) UpdateProject(projectId uint64, title string, description string, tags models.Tags,
	deadline time.Duration, price uint64) error {
	stmtUser, err := s.db.Prepare("UPDATE Projects SET title = $1, description = $2, tags = $3, deadline = $4, price = $5 WHERE id = $6")
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

	_, err = stmtUser.Exec(title, description, tags, deadline, price, projectId)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}

func (s *PgStorage) DeleteProject(projectId uint64) error {
	stmtUser, err := s.db.Prepare("DELETE FROM Projects WHERE id = $1")
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

func (s *PgStorage) CreateBid(projectId uint64, fromUser string, price uint64, deadline time.Duration,
	message string) (uint64, error) {
	stmtBid, err := s.db.Prepare("INSERT INTO Bids (project, user, price, deadline, message) VALUES($1, $2, $3, $4, $5) RETURNING id")
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

	row := stmtBid.QueryRow(projectId, fromUser, price, deadline, message)
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

func (s *PgStorage) GetBid(id uint64) (*models.Bid, error) {
	// TODO: set explicit field set and order in query
	stmtBid, err := s.db.Prepare("SELECT * FROM Bids WHERE id = $1")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return nil, err
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
		return nil, err
	}
	res := models.Bid{}
	err = rowBid.Scan(&res.Id, &res.Project, &res.User, &res.Price, &res.Deadline, &res.Message)
	if err != nil {
		logger.Log.Error("Failed to get data from rowset", zap.Error(err))
		return nil, err
	}
	return &res, nil
}

func (s *PgStorage) GetProjectBids(projectId uint64) ([]uint64, error) {
	stmtBids, err := s.db.Prepare("SELECT id FROM Bids WHERE project = $1")
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
	var res []uint64
	for rowsBids.Next() {
		var id uint64
		err = rowsBids.Scan(&id)
		if err != nil {
			logger.Log.Error("Failed to get data from rowset", zap.Error(err))
			return nil, err
		}
		res = append(res, id)
	}
	err = rowsBids.Err()
	if err != nil {
		logger.Log.Error("error while iterating rows", zap.Error(err))
		return nil, err
	}
	return res, nil
}

func (s *PgStorage) UpdateBid(id uint64, price uint64, deadline time.Duration, message string) error {
	stmtBid, err := s.db.Prepare("UPDATE Bids SET price = $1, deadline = $2, message = $3 WHERE id = $4")
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

	_, err = stmtBid.Exec(price, deadline, message, id)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}

func (s *PgStorage) DeleteBid(id uint64) error {
	stmtBid, err := s.db.Prepare("DELETE FROM Bids WHERE id = $1")
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

func (s *PgStorage) AcceptBid(id uint64) error {
	tx, err := s.db.Begin()
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

	// query to update project data according to bid
	stmtProjectUpdate, err := tx.Prepare("UPDATE Projects SET price = Bids.price, deadline = Bids.deadline, contractor=Bids.user, status = $1 FROM Bids WHERE Projects.id = Bids.project AND Bids.id = $2")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return err
	}
	defer func() {
		err = stmtProjectUpdate.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()
	res, err := stmtProjectUpdate.Exec(models.InWork, id)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get affected rows", zap.Error(err))
		return err
	}
	if rows != 1 {
		logger.Log.Error("invalid affected rows", zap.Int64("rows", rows))
		return err
	}

	// query to delete accepted bid
	stmtDeleteBid, err := tx.Prepare("DELETE FROM Bids WHERE id = $1")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return err
	}
	defer func() {
		err = stmtDeleteBid.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()

	_, err = stmtDeleteBid.Exec(id)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return tx.Commit()
}

func (s *PgStorage) Close() error {
	return s.db.Close()
}
