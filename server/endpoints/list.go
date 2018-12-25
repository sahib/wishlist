package endpoints

import (
	"net/http"

	"github.com/sahib/wedlist/db"
)

type ListResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Items   []*db.Item `json:"items,omitempty"`
}

type ListHandler struct {
	db *db.Database
}

func NewListHandler(db *db.Database) *ListHandler {
	return &ListHandler{db: db}
}

func (lh *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userKey("user")).(*db.User)
	if !ok {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to read user")
		return
	}

	items, err := lh.db.GetItems(user.ID)
	if err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to list from db: %v", err)
		return
	}

	jsonify(w, http.StatusOK, &ListResponse{
		Success: true,
		Items:   items,
	})
}

func (lh *ListHandler) NeedsAuthentication() bool {
	return true
}
