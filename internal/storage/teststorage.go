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

func NewTestStorage(users []models.User, projects []models.Project, bids []models.Bid, rates []models.Rate) *TestStorage {
	return &TestStorage{
		Users:        users,
		Projects:     projects,
		Bids:         bids,
		Rates:        rates,
		MaxProjectId: maxProjectId(projects),
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

func (t *TestStorage) Close() error {
	return nil
}
