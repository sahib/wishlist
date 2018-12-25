package endpoints

import (
	"log"
	"net/http"

	"github.com/sahib/wedlist/cache"
)

type LogoutHandler struct {
	cache *cache.SessionCache
}

func NewLogoutHandler(cache *cache.SessionCache) *LoginHandler {
	return &LoginHandler{cache: cache}
}

func (lh *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "no cookie")
		return
	}

	if err := lh.cache.Forget(cookie.Value); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "deletion failed: %v", err)
		return
	}

	log.Printf("session %s logged out", cookie.Value)
	jsonifyErrf(w, http.StatusOK, "OK")
}

func (lh *LogoutHandler) NeedsAuthentication() bool {
	return true
}
