package storage

import (
	"fmt"
	"sync"

	"github.com/sgladkov/tortuga/internal/models"
)

type TestStorage struct {
	lock     sync.Mutex
	Users    []models.User
	Projects []models.Project
	Bids     []models.Bid
	Rates    []models.Rate
}

func NewTestStorage(users []models.User, projects []models.Project, bids []models.Bid, rates []models.Rate) *TestStorage {
	return &TestStorage{
		Users:    users,
		Projects: projects,
		Bids:     bids,
		Rates:    rates,
	}
}

func (t *TestStorage) GetUserList() (*models.UserList, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	res := models.UserList{}
	for _, u := range t.Users {
		res = append(res, u.Id)
	}
	return &res, nil
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

func (t *TestStorage) GetProjectList() (*models.ProjectList, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	res := models.ProjectList{}
	for _, p := range t.Projects {
		res = append(res, p.Id)
	}
	return &res, nil
}

func (t *TestStorage) GetUserProjects(userId string) (*models.ProjectList, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	res := models.ProjectList{}
	for _, p := range t.Projects {
		if p.Owner == userId {
			res = append(res, p.Id)
		}
	}
	return &res, nil
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

func (t *TestStorage) Close() error {
	return nil
}
