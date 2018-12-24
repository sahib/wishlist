package endpoints

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/sahib/wedlist/cache"
	"github.com/sahib/wedlist/db"
)

type userKey string

type AuthHandler interface {
	NeedsAuthentication() bool
}

type AuthMiddleware struct {
	db    *db.Database
	cache *cache.SessionCache
}

type noAuthEndpoint struct {
	http.Handler
}

func (noe noAuthEndpoint) NeedsAuthentication() bool {
	return false
}

func NoAuth(handler http.Handler) http.Handler {
	return noAuthEndpoint{Handler: handler}
}

func NewAuthMiddleware(db *db.Database, cache *cache.SessionCache) *AuthMiddleware {
	return &AuthMiddleware{db: db, cache: cache}
}

func IsAuthenticated(r *http.Request, cache *cache.SessionCache, db *db.Database) (*db.User, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, fmt.Errorf("failed to extract cookie: %v", err)
	}

	session, err := cache.Session(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to extract session: %v", err)
	}

	if !session.IsConfirmed {
		return nil, nil
	}

	user, err := db.GetUserByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from db: %v", err)
	}

	return user, nil
}

func (amw *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHandler, ok := next.(AuthHandler)
		if !ok || authHandler.NeedsAuthentication() {
			user, err := IsAuthenticated(r, amw.cache, amw.db)
			if err != nil {
				jsonifyErrf(w, 500, "login failed: %v", err)
				return
			}

			if user == nil {
				// Redirect user to login page.
				http.Redirect(w, r, "/login.html", http.StatusSeeOther)
				return
			}

			w.Header().Set("X-CSRF-Token", csrf.Token(r))
			r = r.WithContext(context.WithValue(r.Context(), userKey("user"), user))
		}

		next.ServeHTTP(w, r)
	})
}
