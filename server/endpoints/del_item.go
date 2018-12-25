package endpoints

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jcuga/golongpoll"
	"github.com/sahib/wedlist/db"
)

type DelRequest struct {
	ItemID int64 `json="itemid"`
}

type DelHandler struct {
	db      *db.Database
	pollMgr *golongpoll.LongpollManager
}

func NewDelHandler(db *db.Database, pollMgr *golongpoll.LongpollManager) *DelHandler {
	return &DelHandler{db: db, pollMgr: pollMgr}
}

func (dh *DelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := DelRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json in request body: %v", err)
		return
	}

	log.Printf("DEL REQ: %v", req)
	user, ok := r.Context().Value(userKey("user")).(*db.User)
	if !ok {
		jsonifyErrf(w, http.StatusInternalServerError, "no user in context")
		return
	}

	if err := dh.db.DeleteItem(user.ID, req.ItemID); err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to delete in db: %v", err)
		return
	}

	log.Printf("user %s deleted item %d", user.EMail, req.ItemID)
	if err := dh.pollMgr.Publish("list-change", "del"); err != nil {
		log.Printf("failed to publish event: %v", err)
	}

	jsonifyErrf(w, http.StatusOK, "OK")
}

func (dh *DelHandler) NeedsAuthentication() bool {
	return true
}
