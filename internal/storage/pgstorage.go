package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/sgladkov/tortuga/internal/logger"
	"github.com/sgladkov/tortuga/internal/models"
	"go.uber.org/zap"
	"time"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
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
		"registered bigint, " +
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
		"created bigint, " +
		"status smallint, " +
		"owner varchar(42) REFERENCES Users(id), " +
		"contractor varchar(42) REFERENCES Users(id), " +
		"started bigint, " +
		"deadline bigint, " +
		"price bigint)")
	if err != nil {
		logger.Log.Error("Failed to create exec table", zap.Error(err))
		return err
	}
	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS Bids (" +
		"id bigint PRIMARY KEY generated always as identity, " +
		"project bigint REFERENCES Projects(id), " +
		"contractor varchar(42) REFERENCES Users(id), " +
		"price bigint, " +
		"deadline bigint, " +
		"message text)")
	if err != nil {
		logger.Log.Error("Failed to create exec table", zap.Error(err))
		return err
	}
	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS Rates (" +
		"author varchar(42) REFERENCES Users(id), " +
		"project bigint REFERENCES Projects(id), " +
		"object varchar(42) REFERENCES Users(id), " +
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

func (s *PgStorage) GetUserList(ctx context.Context) ([]models.User, error) {
	rowsUsers, err := s.exec.QueryContext(ctx, "SELECT id, nickname, description, nonce, registered, status, tags, "+
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
		var registered models.FixedTime
		err = rowsUsers.Scan(&user.Id, &user.Nickname, &user.Description, &user.Nonce, &registered,
			&user.Status, &user.Tags, &user.Rating, &user.Account)
		if err != nil {
			logger.Log.Error("Failed to get data from rowset", zap.Error(err))
			return nil, err
		}
		user.Registered = time.Time(registered)
		res = append(res, user)
	}
	err = rowsUsers.Err()
	if err != nil {
		logger.Log.Error("error while iterating rows", zap.Error(err))
		return nil, err
	}
	return res, nil
}

func (s *PgStorage) GetUser(ctx context.Context, id string) (models.User, error) {
	// TODO: set explicit field set and order in query
	stmtUser, err := s.exec.PrepareContext(ctx, "SELECT * FROM Users WHERE id = $1")
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
	rowUser := stmtUser.QueryRowContext(ctx, id)
	if rowUser.Err() != nil {
		logger.Log.Error("Failed to query project data", zap.Error(err))
		return models.User{}, err
	}
	res := models.User{}
	registered := models.FixedTime{}
	err = rowUser.Scan(&res.Id, &res.Nickname, &res.Description, &res.Nonce, &registered, &res.Status,
		&res.Tags, &res.Rating, &res.Account)
	res.Registered = time.Time(registered)
	if err != nil {
		logger.Log.Error("Failed to get data from rowset", zap.Error(err))
		return models.User{}, err
	}
	return res, nil
}

func (s *PgStorage) GetProjectList(ctx context.Context) ([]models.Project, error) {
	rowsProjects, err := s.exec.QueryContext(ctx, "SELECT id, title, description, tags, created, status, owner, "+
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
		var created models.FixedTime
		var contractor sql.NullString
		var started sql.NullInt64
		err = rowsProjects.Scan(&project.Id, &project.Title, &project.Description, &project.Tags,
			&created, &project.Status, &project.Owner, &contractor, &started,
			&project.Deadline, &project.Price)
		if err != nil {
			logger.Log.Error("Failed to get data from rowset", zap.Error(err))
			return nil, err
		}
		project.Created = time.Time(created)
		if contractor.Valid {
			project.Contractor = contractor.String
		}
		if started.Valid {
			project.Started = time.UnixMilli(started.Int64)
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

func (s *PgStorage) GetUserProjects(ctx context.Context, userId string) ([]models.Project, error) {
	stmtProjects, err := s.exec.PrepareContext(ctx, "SELECT id, title, description, tags, created, status, owner, "+
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
	rowsProjects, err := stmtProjects.QueryContext(ctx, userId)
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
		var created models.FixedTime
		var contractor sql.NullString
		var started sql.NullInt64
		err = rowsProjects.Scan(&project.Id, &project.Title, &project.Description, &project.Tags,
			&created, &project.Status, &project.Owner, &contractor, &started,
			&project.Deadline, &project.Price)
		if err != nil {
			logger.Log.Error("Failed to get data from rowset", zap.Error(err))
			return nil, err
		}
		project.Created = time.Time(created)
		if contractor.Valid {
			project.Contractor = contractor.String
		}
		if started.Valid {
			project.Started = time.UnixMilli(started.Int64)
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

func (s *PgStorage) GetProject(ctx context.Context, id uint64) (models.Project, error) {
	// TODO: set explicit field set and order in query
	stmtProject, err := s.exec.PrepareContext(ctx, "SELECT * FROM Projects WHERE id = $1")
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
	rowProject := stmtProject.QueryRowContext(ctx, id)
	if err = rowProject.Err(); err != nil {
		logger.Log.Error("Failed to query project data", zap.Error(err))
		return models.Project{}, err
	}
	res := models.Project{}
	var contractor sql.NullString
	var started sql.NullInt64
	var created models.FixedTime
	err = rowProject.Scan(&res.Id, &res.Title, &res.Description, &res.Tags, &created, &res.Status,
		&res.Owner, &contractor, &started, &res.Deadline, &res.Price)
	if err != nil {
		logger.Log.Error("Failed to get data from rowset", zap.Error(err))
		return models.Project{}, err
	}
	res.Created = time.Time(created)
	if contractor.Valid {
		res.Contractor = contractor.String
	}
	if started.Valid {
		res.Started = time.UnixMilli(started.Int64)
	}
	return res, nil
}

func (s *PgStorage) CreateUser(ctx context.Context, user models.User) error {
	stmtUser, err := s.exec.PrepareContext(ctx, "INSERT INTO Users (id, nickname, description, nonce, registered, "+
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

	registered := models.FixedTime(user.Registered)
	_, err = stmtUser.ExecContext(ctx, user.Id, user.Nickname, user.Description, user.Nonce, registered,
		user.Status, user.Tags, user.Rating, user.Account)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}

func (s *PgStorage) UpdateUser(ctx context.Context, user models.User) error {
	stmtUser, err := s.exec.PrepareContext(ctx, "UPDATE Users SET nickname=$1, description=$2, nonce=$3, registered=$4, "+
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

	registered := models.FixedTime(user.Registered)
	res, err := stmtUser.ExecContext(ctx, user.Nickname, user.Description, user.Nonce, registered, user.Status,
		user.Tags, user.Rating, user.Account, user.Id)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get affected rows", zap.Error(err))
		return err
	}
	if affected != 1 {
		return fmt.Errorf("%d rows affected in update", affected)
	}

	return nil
}

func (s *PgStorage) CreateProject(ctx context.Context, project models.Project) (uint64, error) {
	stmtUser, err := s.exec.PrepareContext(ctx, "INSERT INTO Projects (title, description, tags, created, status, owner, deadline, price) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id")
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

	created := models.FixedTime(project.Created)
	row := stmtUser.QueryRowContext(ctx, project.Title, project.Description, project.Tags, created,
		project.Status, project.Owner, project.Deadline, project.Price)
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

func (s *PgStorage) UpdateProject(ctx context.Context, project models.Project) error {
	stmtUser, err := s.exec.PrepareContext(ctx, "UPDATE Projects SET title=$1, description=$2, tags=$3, created=$4, "+
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

	created := models.FixedTime(project.Created)
	contractor := sql.NullString{
		String: project.Contractor,
		Valid:  project.Contractor != "",
	}
	started := sql.NullInt64{
		Int64: project.Started.UnixMilli(),
		Valid: project.Started != time.Time{},
	}
	res, err := stmtUser.ExecContext(ctx, project.Title, project.Description, project.Tags, created, project.Status,
		project.Owner, contractor, started, project.Deadline, project.Price, project.Id)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get affected rows", zap.Error(err))
		return err
	}
	if affected != 1 {
		return fmt.Errorf("%d rows affected in update", affected)
	}

	return nil
}

func (s *PgStorage) DeleteUser(ctx context.Context, id string) error {
	stmtUser, err := s.exec.PrepareContext(ctx, "DELETE FROM Users WHERE id = $1")
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

	res, err := stmtUser.ExecContext(ctx, id)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get affected rows", zap.Error(err))
		return err
	}
	if affected != 1 {
		return fmt.Errorf("%d rows affected in delete", affected)
	}

	return nil
}

func (s *PgStorage) DeleteProject(ctx context.Context, projectId uint64) error {
	stmtUser, err := s.exec.PrepareContext(ctx, "DELETE FROM Projects WHERE id = $1")
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
	res, err := stmtUser.ExecContext(ctx, projectId)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get affected rows", zap.Error(err))
		return err
	}
	if affected != 1 {
		return fmt.Errorf("%d rows affected in delete", affected)
	}

	return nil
}

func (s *PgStorage) CreateBid(ctx context.Context, bid models.Bid) (uint64, error) {
	stmtBid, err := s.exec.PrepareContext(ctx, "INSERT INTO Bids (project, contractor, price, deadline, message) VALUES($1, $2, $3, $4, $5) RETURNING id")
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

	row := stmtBid.QueryRowContext(ctx, bid.Project, bid.User, bid.Price, bid.Deadline, bid.Message)
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

func (s *PgStorage) GetBid(ctx context.Context, id uint64) (models.Bid, error) {
	// TODO: set explicit field set and order in query
	stmtBid, err := s.exec.PrepareContext(ctx, "SELECT * FROM Bids WHERE id = $1")
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
	rowBid := stmtBid.QueryRowContext(ctx, id)
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

func (s *PgStorage) GetProjectBids(ctx context.Context, projectId uint64) ([]models.Bid, error) {
	stmtBids, err := s.exec.PrepareContext(ctx, "SELECT id, project, contractor, price, deadline, message FROM Bids WHERE project = $1")
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
	rowsBids, err := stmtBids.QueryContext(ctx, projectId)
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

func (s *PgStorage) UpdateBid(ctx context.Context, bid models.Bid) error {
	stmtBid, err := s.exec.PrepareContext(ctx, "UPDATE Bids SET project=$1, contractor=$2, price=$3, deadline = $4, message = $5 WHERE id = $6")
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

	_, err = stmtBid.ExecContext(ctx, bid.Project, bid.User, bid.Price, bid.Deadline, bid.Message, bid.Id)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}

func (s *PgStorage) DeleteBid(ctx context.Context, id uint64) error {
	stmtBid, err := s.exec.PrepareContext(ctx, "DELETE FROM Bids WHERE id = $1")
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

	_, err = stmtBid.ExecContext(ctx, id)
	if err != nil {
		logger.Log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}

func (s *PgStorage) Close() error {
	return s.db.Close()
}
