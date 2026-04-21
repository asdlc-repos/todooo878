package store

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/todooo/todo-api/internal/models"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrEmailExists   = errors.New("email already exists")
	ErrForbidden     = errors.New("forbidden")
)

type Store interface {
	// Users
	CreateUser(email, passwordHash string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id string) (*models.User, error)

	// Categories
	CreateCategory(userID, name string) (*models.Category, error)
	ListCategories(userID string) []*models.Category
	GetCategory(id string) (*models.Category, error)
	DeleteCategory(userID, id string) error

	// Tasks
	CreateTask(task *models.Task) (*models.Task, error)
	GetTask(id string) (*models.Task, error)
	ListTasks(userID string) []*models.Task
	UpdateTask(task *models.Task) (*models.Task, error)
	DeleteTask(userID, id string) error
}

type Memory struct {
	mu         sync.RWMutex
	users      map[string]*models.User // id -> user
	emailIndex map[string]string       // email -> id
	categories map[string]*models.Category
	tasks      map[string]*models.Task
}

func NewMemory() *Memory {
	return &Memory{
		users:      make(map[string]*models.User),
		emailIndex: make(map[string]string),
		categories: make(map[string]*models.Category),
		tasks:      make(map[string]*models.Task),
	}
}

func normalizeEmail(e string) string {
	return strings.ToLower(strings.TrimSpace(e))
}

func (m *Memory) CreateUser(email, passwordHash string) (*models.User, error) {
	email = normalizeEmail(email)
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.emailIndex[email]; ok {
		return nil, ErrEmailExists
	}
	u := &models.User{
		ID:           newID(),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}
	m.users[u.ID] = u
	m.emailIndex[email] = u.ID
	return u, nil
}

func (m *Memory) GetUserByEmail(email string) (*models.User, error) {
	email = normalizeEmail(email)
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.emailIndex[email]
	if !ok {
		return nil, ErrNotFound
	}
	u, ok := m.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (m *Memory) GetUserByID(id string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (m *Memory) CreateCategory(userID, name string) (*models.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := &models.Category{
		ID:        newID(),
		UserID:    userID,
		Name:      strings.TrimSpace(name),
		CreatedAt: time.Now().UTC(),
	}
	m.categories[c.ID] = c
	return c, nil
}

func (m *Memory) ListCategories(userID string) []*models.Category {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*models.Category, 0)
	for _, c := range m.categories {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out
}

func (m *Memory) GetCategory(id string) (*models.Category, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.categories[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}

func (m *Memory) DeleteCategory(userID, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.categories[id]
	if !ok {
		return ErrNotFound
	}
	if c.UserID != userID {
		return ErrForbidden
	}
	delete(m.categories, id)
	// Unset categoryId on tasks that referenced it (still owned by the user)
	for _, t := range m.tasks {
		if t.CategoryID != nil && *t.CategoryID == id {
			t.CategoryID = nil
			t.UpdatedAt = time.Now().UTC()
		}
	}
	return nil
}

func (m *Memory) CreateTask(task *models.Task) (*models.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	task.ID = newID()
	now := time.Now().UTC()
	task.CreatedAt = now
	task.UpdatedAt = now
	m.tasks[task.ID] = task
	return task, nil
}

func (m *Memory) GetTask(id string) (*models.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tasks[id]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

func (m *Memory) ListTasks(userID string) []*models.Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*models.Task, 0)
	for _, t := range m.tasks {
		if t.UserID == userID {
			out = append(out, t)
		}
	}
	return out
}

func (m *Memory) UpdateTask(task *models.Task) (*models.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.tasks[task.ID]
	if !ok {
		return nil, ErrNotFound
	}
	task.CreatedAt = existing.CreatedAt
	task.UpdatedAt = time.Now().UTC()
	m.tasks[task.ID] = task
	return task, nil
}

func (m *Memory) DeleteTask(userID, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return ErrNotFound
	}
	if t.UserID != userID {
		return ErrForbidden
	}
	delete(m.tasks, id)
	return nil
}
