package storage

import (
	"fmt"
	"github.com/sgladkov/tortuga/internal/models"
	"sync"
	"time"
)

type TestStorage struct {
	lock         sync.Mutex
	Users        []models.User
	Projects     []models.Project
	Bids         []models.Bid
	Rates        []models.Rate
	MaxProjectId uint64
	MaxBidId     uint64
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

func NewTestStorage(users []models.User, projects []models.Project, bids []models.Bid, rates []models.Rate) *TestStorage {
	return &TestStorage{
		Users:        users,
		Projects:     projects,
		Bids:         bids,
		Rates:        rates,
		MaxProjectId: maxProjectId(projects),
		MaxBidId:     maxBidId(bids),
	}
}

func (t *TestStorage) BeginTx() error {
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
	}
	return nil
}

func (t *TestStorage) CommitTx() error {
	if t.dataCopy == nil {
		return fmt.Errorf("transaction is closed already")
	}
	t.dataCopy = nil
	return nil
}

func (t *TestStorage) RollbackTx() error {
	if t.dataCopy == nil {
		return fmt.Errorf("transaction is closed already")
	}
	t.Users = t.dataCopy.Users
	t.Projects = t.dataCopy.Projects
	t.Bids = t.dataCopy.Bids
	t.Rates = t.dataCopy.Rates
	t.MaxProjectId = t.dataCopy.MaxProjectId
	t.MaxBidId = t.dataCopy.MaxBidId
	t.dataCopy = nil
	return nil
}

func (t *TestStorage) GetUserList() ([]models.User, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	var res []models.User
	for _, u := range t.Users {
		res = append(res, u)
	}
	return res, nil
}

func (t *TestStorage) GetUser(id string) (*models.User, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, u := range t.Users {
		if u.Id == id {
			return &u, nil
		}
	}
	return nil, fmt.Errorf("no user with id %s", id)
}

func (t *TestStorage) GetProjectList() ([]models.Project, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	var res []models.Project
	for _, p := range t.Projects {
		res = append(res, p)
	}
	return res, nil
}

func (t *TestStorage) GetUserProjects(userId string) ([]models.Project, error) {
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

func (t *TestStorage) GetProject(id uint64) (*models.Project, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, p := range t.Projects {
		if p.Id == id {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("no project with id %v", id)
}

func (t *TestStorage) AddUser(user *models.User) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.Users = append(t.Users, *user)
	return nil
}

func (t *TestStorage) UpdateUserNonce(id string, nonce uint64) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, u := range t.Users {
		if u.Id == id {
			u.Nonce = nonce
			return nil
		}
	}
	return fmt.Errorf("no user with id %s", id)
}

func (t *TestStorage) CreateProject(title string, description string, tags models.Tags, owner string, deadline time.Duration, price uint64) (uint64, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	project := models.Project{
		Id:          t.MaxProjectId + 1,
		Title:       title,
		Description: description,
		Tags:        tags,
		Created:     time.Now(),
		Status:      models.Open,
		Owner:       owner,
		Deadline:    deadline,
		Price:       price,
	}
	t.Projects = append(t.Projects, project)
	t.MaxProjectId++
	return project.Id, nil
}

func (t *TestStorage) UpdateProject(projectId uint64, title string, description string, tags models.Tags,
	deadline time.Duration, price uint64) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, p := range t.Projects {
		if p.Id == projectId {
			p.Title = title
			p.Description = description
			p.Tags = tags
			p.Deadline = deadline
			p.Price = price
			return nil
		}
	}
	return fmt.Errorf("no project with id %d", projectId)
}

func (t *TestStorage) DeleteProject(projectId uint64) error {
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

func (t *TestStorage) CreateBid(projectId uint64, fromUser string, price uint64, deadline time.Duration,
	message string) (uint64, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	found := false
	for _, p := range t.Projects {
		if p.Id == projectId {
			found = true
		}
	}
	if !found {
		return 0, fmt.Errorf("no project with id %d", projectId)
	}
	found = false
	for _, u := range t.Users {
		if u.Id == fromUser {
			found = true
		}
	}
	if !found {
		return 0, fmt.Errorf("no user with id %d", fromUser)
	}

	bid := models.Bid{
		Id:       t.MaxBidId + 1,
		Project:  projectId,
		User:     fromUser,
		Deadline: deadline,
		Price:    price,
		Message:  message,
	}
	t.Bids = append(t.Bids, bid)
	t.MaxBidId++
	return bid.Id, nil
}

func (t *TestStorage) GetBid(id uint64) (*models.Bid, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, b := range t.Bids {
		if b.Id == id {
			return &b, nil
		}
	}
	return nil, fmt.Errorf("no bid with id %v", id)
}

func (t *TestStorage) GetProjectBids(projectId uint64) ([]models.Bid, error) {
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

func (t *TestStorage) UpdateBid(id uint64, price uint64, deadline time.Duration, message string) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, b := range t.Bids {
		if b.Id == id {
			b.Deadline = deadline
			b.Price = price
			b.Message = message
			return nil
		}
	}
	return fmt.Errorf("no bid with id %d", id)
}

func (t *TestStorage) DeleteBid(id uint64) error {
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

func (t *TestStorage) AcceptBid(id uint64) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for idx, b := range t.Bids {
		if b.Id == id {
			found := false
			for _, p := range t.Projects {
				if p.Id == b.Project {
					p.Contractor = b.User
					p.Deadline = b.Deadline
					p.Price = b.Price
					p.Status = models.InWork
					found = true
				}
			}
			if !found {
				return fmt.Errorf("no project with id %d", b.Project)
			}
			t.Bids = append(t.Bids[:idx], t.Bids[idx+1:]...)
			return nil
		}
	}
	return fmt.Errorf("no bid with id %d", id)
}

func (t *TestStorage) CancelProject(id uint64) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, p := range t.Projects {
		if p.Id == id {
			if p.Status == models.InWork || p.Status == models.InReview {
				p.Status = models.Canceled
				return nil
			}
			return fmt.Errorf("invalid project status %d", p.Status)
		}
	}
	return fmt.Errorf("no project with id %d", id)
}

func (t *TestStorage) SetProjectReady(id uint64) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, p := range t.Projects {
		if p.Id == id {
			if p.Status == models.InWork {
				p.Status = models.InReview
				return nil
			}
			return fmt.Errorf("invalid project status %d", p.Status)
		}
	}
	return fmt.Errorf("no project with id %d", id)
}

func (t *TestStorage) AcceptProject(id uint64) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, p := range t.Projects {
		if p.Id == id {
			if p.Status == models.InReview {
				p.Status = models.Completed
				return nil
			}
			return fmt.Errorf("invalid project status %d", p.Status)
		}
	}
	return fmt.Errorf("no project with id %d", id)
}

func (t *TestStorage) Close() error {
	return nil
}
