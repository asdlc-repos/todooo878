package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/todooo/todo-api/internal/httpx"
	"github.com/todooo/todo-api/internal/middleware"
	"github.com/todooo/todo-api/internal/store"
)

type CategoryHandler struct {
	Store store.Store
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserID(r)
	cats := h.Store.ListCategories(uid)
	httpx.WriteJSON(w, http.StatusOK, cats)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserID(r)
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if len(name) > 100 {
		httpx.WriteError(w, http.StatusBadRequest, "name is too long")
		return
	}
	c, err := h.Store.CreateCategory(uid, name)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not create category")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, c)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request, id string) {
	uid, _ := middleware.UserID(r)
	err := h.Store.DeleteCategory(uid, id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			httpx.WriteError(w, http.StatusNotFound, "category not found")
		case errors.Is(err, store.ErrForbidden):
			httpx.WriteError(w, http.StatusForbidden, "forbidden")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "could not delete category")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
