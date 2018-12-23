package endpoints

import (
	"encoding/json"
	"html"
	"net/http"

	"github.com/sahib/wedlist/db"
)

type AddRequest struct {
	Name string `json="name"`
	Link string `json="link"`
}

type AddHandler struct {
	db *db.Database
}

func NewAddHandler(db *db.Database) *AddHandler {
	return &AddHandler{db: db}
}

func (ah *AddHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := AddRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json in request body: %v", err)
		return
	}

	user, ok := r.Context().Value(userKey("user")).(*db.User)
	if !ok {
		jsonifyErrf(w, http.StatusInternalServerError, "no user in context")
		return
	}

	req.Name = html.EscapeString(req.Name)
	if _, err := ah.db.AddItem(req.Name, req.Link, user.ID); err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to add to database: %v", err)
		return
	}

	jsonifyErrf(w, http.StatusCreated, "OK")
}

func (ah *AddHandler) NeedsAuthentication() bool {
	return true
}
