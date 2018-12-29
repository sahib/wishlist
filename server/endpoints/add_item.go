package endpoints

import (
	"encoding/json"
	"html"
	"log"
	"net/http"

	"github.com/jcuga/golongpoll"
	"github.com/sahib/wishlist/db"
)

type AddRequest struct {
	Name string `json="name"`
	Link string `json="link"`
}

type AddHandler struct {
	db      *db.Database
	pollMgr *golongpoll.LongpollManager
}

func NewAddHandler(db *db.Database, pollMgr *golongpoll.LongpollManager) *AddHandler {
	return &AddHandler{db: db, pollMgr: pollMgr}
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
	if _, err := ah.db.AddItem(req.Name, req.Link, user.ID, user.ID); err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to add to database: %v", err)
		return
	}

	log.Printf("add new item: %s", req.Name)
	if err := ah.pollMgr.Publish("list-change", "add"); err != nil {
		log.Printf("failed to publish event: %v", err)
	}

	jsonifyErrf(w, http.StatusCreated, "OK")
}

func (ah *AddHandler) NeedsAuthentication() bool {
	return true
}
