package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/todooo/todo-api/internal/auth"
	"github.com/todooo/todo-api/internal/httpx"
	"github.com/todooo/todo-api/internal/middleware"
	"github.com/todooo/todo-api/internal/store"
)

type AuthHandler struct {
	Store    store.Store
	Auth     *auth.Manager
	CookieSecure bool
	CookieDomain string
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	c.Email = strings.TrimSpace(c.Email)
	if _, err := mail.ParseAddress(c.Email); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid email")
		return
	}
	if len(c.Password) < 8 {
		httpx.WriteError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bcrypt error: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "could not create user")
		return
	}
	u, err := h.Store.CreateUser(c.Email, string(hash))
	if err != nil {
		if errors.Is(err, store.ErrEmailExists) {
			httpx.WriteError(w, http.StatusConflict, "email already exists")
			return
		}
		log.Printf("create user error: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "could not create user")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":    u.ID,
		"email": u.Email,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	c.Email = strings.TrimSpace(c.Email)
	u, err := h.Store.GetUserByEmail(c.Email)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(c.Password)); err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, exp, err := h.Auth.IssueToken(u.ID)
	if err != nil {
		log.Printf("issue token error: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "could not create session")
		return
	}
	csrf := auth.RandomToken(32)
	h.setSessionCookie(w, token, exp)
	h.setCSRFCookie(w, csrf, exp)
	httpx.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":        u.ID,
		"email":     u.Email,
		"csrfToken": csrf,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearCookie(w, auth.SessionCookie, true)
	h.clearCookie(w, auth.CSRFCookie, false)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserID(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	u, err := h.Store.GetUserByID(uid)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":    u.ID,
		"email": u.Email,
	})
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, token string, exp time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookie,
		Value:    token,
		Path:     "/",
		Domain:   h.CookieDomain,
		Expires:  exp,
		MaxAge:   int(time.Until(exp).Seconds()),
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) setCSRFCookie(w http.ResponseWriter, token string, exp time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CSRFCookie,
		Value:    token,
		Path:     "/",
		Domain:   h.CookieDomain,
		Expires:  exp,
		MaxAge:   int(time.Until(exp).Seconds()),
		HttpOnly: false,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) clearCookie(w http.ResponseWriter, name string, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Domain:   h.CookieDomain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: httpOnly,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}
