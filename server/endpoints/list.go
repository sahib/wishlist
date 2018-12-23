package endpoints

import (
	"net/http"

	"github.com/sahib/wedlist/db"
)

type ListItem struct {
	*db.Item
	IsReserved bool `json:"is_reserved"`
}

type ListResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Items   []ListItem `json:"items,omitempty"`
}

type ListHandler struct {
	db *db.Database
}

func NewListHandler(db *db.Database) *ListHandler {
	return &ListHandler{db: db}
}

func (lh *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*db.User)
	if !ok {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to read user")
		return
	}

	items, err := lh.db.GetItems(user.ID)
	if err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to list from db: %v", err)
		return
	}

	lsi := []ListItem{}
	for _, item := range items {
		resUserID, err := lh.db.GetUserForReservation(item.ID)
		if err != nil {
			jsonifyErrf(
				w,
				http.StatusInternalServerError,
				"failed to get reservation from db: %v", err,
			)
			return
		}

		lsi = append(lsi, ListItem{
			Item:       item,
			IsReserved: resUserID >= 0,
		})
	}

	jsonify(w, http.StatusOK, &ListResponse{
		Success: true,
		Items:   lsi,
	})
}

func (lh *ListHandler) NeedsAuthentication() bool {
	return true
}
