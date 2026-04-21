package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/todooo/todo-api/internal/httpx"
	"github.com/todooo/todo-api/internal/middleware"
	"github.com/todooo/todo-api/internal/models"
	"github.com/todooo/todo-api/internal/store"
)

const maxTasks = 1000

type TaskHandler struct {
	Store store.Store
}

type taskCreateBody struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	DueDate     *string `json:"dueDate"`
	CategoryID  *string `json:"categoryId"`
}

type taskUpdateBody struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	DueDate     *string `json:"dueDate"`
	// distinguish "categoryId not provided" from "categoryId=null"
	CategoryIDRaw json.RawMessage `json:"categoryId"`
	Completed     *bool           `json:"completed"`
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserID(r)
	q := r.URL.Query()
	category := q.Get("category")
	status := q.Get("status")
	if status == "" {
		status = "all"
	}
	sortOrder := q.Get("sort")

	tasks := h.Store.ListTasks(uid)
	filtered := tasks[:0]
	for _, t := range tasks {
		if category != "" {
			if t.CategoryID == nil || *t.CategoryID != category {
				continue
			}
		}
		switch status {
		case "active":
			if t.Completed {
				continue
			}
		case "completed":
			if !t.Completed {
				continue
			}
		case "all", "":
		default:
			httpx.WriteError(w, http.StatusBadRequest, "invalid status filter")
			return
		}
		filtered = append(filtered, t)
	}

	switch sortOrder {
	case "", "due_asc":
		sort.SliceStable(filtered, func(i, j int) bool {
			return dueLess(filtered[i].DueDate, filtered[j].DueDate, true)
		})
	case "due_desc":
		sort.SliceStable(filtered, func(i, j int) bool {
			return dueLess(filtered[i].DueDate, filtered[j].DueDate, false)
		})
	default:
		httpx.WriteError(w, http.StatusBadRequest, "invalid sort")
		return
	}

	if len(filtered) > maxTasks {
		filtered = filtered[:maxTasks]
	}
	httpx.WriteJSON(w, http.StatusOK, filtered)
}

// dueLess sorts items so that non-nil due dates come first (ascending/descending)
// and nil dates sort last in both orders.
func dueLess(a, b *time.Time, asc bool) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil {
		return false
	}
	if b == nil {
		return true
	}
	if asc {
		return a.Before(*b)
	}
	return a.After(*b)
}

func parseDueDate(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, errors.New("dueDate must be YYYY-MM-DD")
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	today := time.Now().UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	maxDate := today.AddDate(10, 0, 0)
	if t.Before(today) {
		return nil, errors.New("dueDate must be today or later")
	}
	if t.After(maxDate) {
		return nil, errors.New("dueDate must be within 10 years from today")
	}
	return &t, nil
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserID(r)
	var body taskCreateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		httpx.WriteError(w, http.StatusBadRequest, "title is required")
		return
	}
	if len(title) > 200 {
		httpx.WriteError(w, http.StatusBadRequest, "title is too long")
		return
	}
	var due *time.Time
	if body.DueDate != nil && *body.DueDate != "" {
		d, err := parseDueDate(*body.DueDate)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		due = d
	}
	var catID *string
	if body.CategoryID != nil && *body.CategoryID != "" {
		c, err := h.Store.GetCategory(*body.CategoryID)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "category not found")
			return
		}
		if c.UserID != uid {
			httpx.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}
		catID = &c.ID
	}
	t := &models.Task{
		UserID:      uid,
		Title:       title,
		Description: strings.TrimSpace(body.Description),
		DueDate:     due,
		CategoryID:  catID,
		Completed:   false,
	}
	created, err := h.Store.CreateTask(t)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not create task")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, created)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request, id string) {
	uid, _ := middleware.UserID(r)
	t, err := h.Store.GetTask(id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "task not found")
		return
	}
	if t.UserID != uid {
		httpx.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, t)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request, id string) {
	uid, _ := middleware.UserID(r)
	t, err := h.Store.GetTask(id)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "task not found")
		return
	}
	if t.UserID != uid {
		httpx.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	var body taskUpdateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	updated := *t
	if body.Title != nil {
		title := strings.TrimSpace(*body.Title)
		if title == "" {
			httpx.WriteError(w, http.StatusBadRequest, "title is required")
			return
		}
		if len(title) > 200 {
			httpx.WriteError(w, http.StatusBadRequest, "title is too long")
			return
		}
		updated.Title = title
	}
	if body.Description != nil {
		updated.Description = strings.TrimSpace(*body.Description)
	}
	if body.DueDate != nil {
		if *body.DueDate == "" {
			updated.DueDate = nil
		} else {
			d, err := parseDueDate(*body.DueDate)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
			updated.DueDate = d
		}
	}
	if len(body.CategoryIDRaw) > 0 {
		raw := strings.TrimSpace(string(body.CategoryIDRaw))
		if raw == "null" {
			updated.CategoryID = nil
		} else {
			var cid string
			if err := json.Unmarshal(body.CategoryIDRaw, &cid); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "invalid categoryId")
				return
			}
			cid = strings.TrimSpace(cid)
			if cid == "" {
				updated.CategoryID = nil
			} else {
				c, err := h.Store.GetCategory(cid)
				if err != nil {
					httpx.WriteError(w, http.StatusBadRequest, "category not found")
					return
				}
				if c.UserID != uid {
					httpx.WriteError(w, http.StatusForbidden, "forbidden")
					return
				}
				updated.CategoryID = &c.ID
			}
		}
	}
	if body.Completed != nil {
		updated.Completed = *body.Completed
	}
	saved, err := h.Store.UpdateTask(&updated)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not update task")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, saved)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request, id string) {
	uid, _ := middleware.UserID(r)
	if err := h.Store.DeleteTask(uid, id); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			httpx.WriteError(w, http.StatusNotFound, "task not found")
		case errors.Is(err, store.ErrForbidden):
			httpx.WriteError(w, http.StatusForbidden, "forbidden")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "could not delete task")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
