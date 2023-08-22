package storage

import (
	"context"
	"fmt"
	"github.com/sgladkov/tortuga/internal/models"
	"sync"
)

type TestStorage struct {
	lock         sync.Mutex
	Users        []models.User
	Projects     []models.Project
	Bids         []models.Bid
	Rates        []models.Rate
	MaxProjectId uint64
	MaxBidId     uint64
	MaxRateId    uint64
	dataCopy     *TestStorage
}

func maxProjectId(projects []models.Project) uint64 {
	res := uint64(0)
	for _, p := range projects {
		if p.Id > res {
			res = p.Id
		}
	}
	return res
}

func maxBidId(bids []models.Bid) uint64 {
	res := uint64(0)
	for _, b := range bids {
		if b.Id > res {
			res = b.Id
		}
	}
	return res
}

func maxRateId(rates []models.Rate) uint64 {
	res := uint64(0)
	for _, r := range rates {
		if r.Id > res {
			res = r.Id
		}
	}
	return res
}

func NewTestStorage(users []models.User, projects []models.Project, bids []models.Bid, rates []models.Rate) *TestStorage {
	return &TestStorage{
		Users:        users,
		Projects:     projects,
		Bids:         bids,
		Rates:        rates,
		MaxProjectId: maxProjectId(projects),
		MaxBidId:     maxBidId(bids),
		MaxRateId:    maxRateId(rates),
	}
}

func (t *TestStorage) BeginTx() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.dataCopy != nil {
		return fmt.Errorf("transaction is open already")
	}
	t.dataCopy = &TestStorage{
		Users:        t.Users,
		Projects:     t.Projects,
		Bids:         t.Bids,
		Rates:        t.Rates,
		MaxProjectId: t.MaxProjectId,
		MaxBidId:     t.MaxBidId,
		MaxRateId:    t.MaxRateId,
	}
	return nil
}

func (t *TestStorage) CommitTx() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.dataCopy == nil {
		return fmt.Errorf("transaction is closed already")
	}
	t.dataCopy = nil
	return nil
}

func (t *TestStorage) RollbackTx() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.dataCopy == nil {
		return fmt.Errorf("transaction is closed already")
	}
	t.Users = t.dataCopy.Users
	t.Projects = t.dataCopy.Projects
	t.Bids = t.dataCopy.Bids
	t.Rates = t.dataCopy.Rates
	t.MaxProjectId = t.dataCopy.MaxProjectId
	t.MaxBidId = t.dataCopy.MaxBidId
	t.MaxRateId = t.dataCopy.MaxRateId
	t.dataCopy = nil
	return nil
}

func (t *TestStorage) GetUserList(_ context.Context) ([]models.User, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	var res []models.User
	for _, u := range t.Users {
		res = append(res, u)
	}
	return res, nil
}

func (t *TestStorage) GetUser(_ context.Context, id string) (models.User, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, u := range t.Users {
		if u.Id == id {
			return u, nil
		}
	}
	return models.User{}, fmt.Errorf("no user with id %s", id)
}

func (t *TestStorage) DeleteUser(_ context.Context, id string) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for idx, u := range t.Users {
		if u.Id == id {
			t.Users = append(t.Users[:idx], t.Users[idx+1:]...)
			return nil
		}
	}
	return fmt.Errorf("no user with id %s", id)
}

func (t *TestStorage) GetProjectList(_ context.Context) ([]models.Project, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	var res []models.Project
	for _, p := range t.Projects {
		res = append(res, p)
	}
	return res, nil
}

func (t *TestStorage) GetUserProjects(_ context.Context, userId string) ([]models.Project, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	var res []models.Project
	for _, p := range t.Projects {
		if p.Owner == userId {
			res = append(res, p)
		}
	}
	return res, nil
}

func (t *TestStorage) GetProject(_ context.Context, id uint64) (models.Project, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, p := range t.Projects {
		if p.Id == id {
			return p, nil
		}
	}
	return models.Project{}, fmt.Errorf("no project with id %v", id)
}

func (t *TestStorage) CreateUser(_ context.Context, user models.User) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, u := range t.Users {
		if u.Id == user.Id {
			return fmt.Errorf("user %s already exists", user.Id)
		}
	}
	t.Users = append(t.Users, user)
	return nil
}

func (t *TestStorage) UpdateUser(_ context.Context, user models.User) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for idx, u := range t.Users {
		if u.Id == user.Id {
			t.Users[idx] = user
			return nil
		}
	}
	return fmt.Errorf("no user with id %v", user.Id)
}

func (t *TestStorage) CreateProject(_ context.Context, project models.Project) (uint64, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, u := range t.Users {
		if u.Id == project.Owner {
			project.Id = t.MaxProjectId + 1
			t.Projects = append(t.Projects, project)
			t.MaxProjectId++
			return project.Id, nil
		}
	}
	return 0, fmt.Errorf("no user %s", project.Owner)
}

func (t *TestStorage) UpdateProject(_ context.Context, project models.Project) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for idx, p := range t.Projects {
		if p.Id == project.Id {
			t.Projects[idx] = project
			return nil
		}
	}
	return fmt.Errorf("no project with id %d", project.Id)
}

func (t *TestStorage) DeleteProject(_ context.Context, projectId uint64) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for idx, p := range t.Projects {
		if p.Id == projectId {
			t.Projects = append(t.Projects[:idx], t.Projects[idx+1:]...)
			return nil
		}
	}
	return fmt.Errorf("no project with id %d", projectId)
}

func (t *TestStorage) CreateBid(_ context.Context, bid models.Bid) (uint64, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	found := false
	for _, p := range t.Projects {
		if p.Id == bid.Project {
			found = true
		}
	}
	if !found {
		return 0, fmt.Errorf("no project with id %v", bid.Project)
	}
	found = false
	for _, u := range t.Users {
		if u.Id == bid.User {
			found = true
		}
	}
	if !found {
		return 0, fmt.Errorf("no user with id %v", bid.User)
	}

	bid.Id = t.MaxBidId + 1
	t.Bids = append(t.Bids, bid)
	t.MaxBidId++
	return bid.Id, nil
}

func (t *TestStorage) GetBid(_ context.Context, id uint64) (models.Bid, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, b := range t.Bids {
		if b.Id == id {
			return b, nil
		}
	}
	return models.Bid{}, fmt.Errorf("no bid with id %v", id)
}

func (t *TestStorage) GetProjectBids(_ context.Context, projectId uint64) ([]models.Bid, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	var res []models.Bid
	for _, b := range t.Bids {
		if b.Project == projectId {
			res = append(res, b)
		}
	}
	return res, nil
}

func (t *TestStorage) UpdateBid(_ context.Context, bid models.Bid) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, b := range t.Bids {
		if b.Id == bid.Id {
			b = bid
			return nil
		}
	}
	return fmt.Errorf("no bid with id %v", bid.Id)
}

func (t *TestStorage) DeleteBid(_ context.Context, id uint64) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for idx, b := range t.Bids {
		if b.Id == id {
			t.Bids = append(t.Bids[:idx], t.Bids[idx+1:]...)
			return nil
		}
	}
	return fmt.Errorf("no bid with id %d", id)
}

func (t *TestStorage) CreateRate(ctx context.Context, rate models.Rate) (uint64, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	found := false
	for _, p := range t.Projects {
		if p.Id == rate.Project {
			found = true
		}
	}
	if !found {
		return 0, fmt.Errorf("no project with id %v", rate.Project)
	}
	found = false
	for _, u := range t.Users {
		if u.Id == rate.Evaluator {
			found = true
		}
	}
	if !found {
		return 0, fmt.Errorf("no user with id %v", rate.Evaluator)
	}
	found = false
	for _, u := range t.Users {
		if u.Id == rate.Evaluated {
			found = true
		}
	}
	if !found {
		return 0, fmt.Errorf("no user with id %v", rate.Evaluated)
	}

	rate.Id = t.MaxRateId + 1
	t.Rates = append(t.Rates, rate)
	t.MaxRateId++
	return rate.Id, nil
}

func (t *TestStorage) GetRate(ctx context.Context, id uint64) (models.Rate, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, r := range t.Rates {
		if r.Id == id {
			return r, nil
		}
	}
	return models.Rate{}, fmt.Errorf("no rate with id %v", id)
}

func (t *TestStorage) GetEvaluatorRates(ctx context.Context, userId string) ([]models.Rate, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	var res []models.Rate
	for _, r := range t.Rates {
		if r.Evaluator == userId {
			res = append(res, r)
		}
	}
	return res, nil
}

func (t *TestStorage) GetEvaluatedRates(ctx context.Context, userId string) ([]models.Rate, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	var res []models.Rate
	for _, r := range t.Rates {
		if r.Evaluated == userId {
			res = append(res, r)
		}
	}
	return res, nil
}

func (t *TestStorage) DeleteRate(ctx context.Context, id uint64) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for idx, r := range t.Rates {
		if r.Id == id {
			t.Rates = append(t.Rates[:idx], t.Rates[idx+1:]...)
			return nil
		}
	}
	return fmt.Errorf("no bid with id %d", id)
}

func (t *TestStorage) Close() error {
	return nil
}
