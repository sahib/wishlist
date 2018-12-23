package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/sahib/wedlist/db"
)

type DelRequest struct {
	ItemID int64 `json="itemid"`
}

type DelHandler struct {
	db *db.Database
}

func NewDelHandler(db *db.Database) *DelHandler {
	return &DelHandler{db: db}
}

func (dh *DelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := DelRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json in request body: %v", err)
		return
	}

	user, ok := r.Context().Value(userKey("user")).(*db.User)
	if !ok {
		jsonifyErrf(w, http.StatusInternalServerError, "no user in context")
		return
	}

	if err := dh.db.DeleteItem(user.ID, req.ItemID); err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to delete in db: %v", err)
		return
	}

	jsonifyErrf(w, http.StatusOK, "OK")
}

func (dh *DelHandler) NeedsAuthentication() bool {
	return true
}
