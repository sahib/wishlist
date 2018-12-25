package endpoints

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jcuga/golongpoll"
	"github.com/sahib/wedlist/db"
)

type ReserveRequest struct {
	ItemID    int64 `json:"item_id"`
	DoReserve bool  `json:"do_reserve"`
}

type ReserveHandler struct {
	db      *db.Database
	pollMgr *golongpoll.LongpollManager
}

func NewReserveHandler(db *db.Database, pollMgr *golongpoll.LongpollManager) *ReserveHandler {
	return &ReserveHandler{db: db, pollMgr: pollMgr}
}

func (rh *ReserveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userKey("user")).(*db.User)
	if !ok {
		jsonifyErrf(w, http.StatusInternalServerError, "no user in context")
		return
	}

	req := ReserveRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json in request body: %v", err)
		return
	}

	resUserID, err := rh.db.GetUserForReservation(req.ItemID)
	if err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to check reservation: %v", err)
		return
	}

	isReserved := resUserID >= 0
	if req.DoReserve {
		if isReserved {
			jsonifyErrf(w, http.StatusUnauthorized, "already reserved")
			return
		}

		log.Printf("reserving item %d for user %s", req.ItemID, user.EMail)
		if err := rh.db.Reserve(user.ID, req.ItemID); err != nil {
			jsonifyErrf(w, http.StatusInternalServerError, "failed to reserve: %v", err)
			return
		}

		if err := rh.pollMgr.Publish("list-change", "reserve"); err != nil {
			log.Printf("failed to publish event: %v", err)
		}
	} else {
		if !isReserved {
			jsonifyErrf(w, http.StatusUnauthorized, "not reserved yet")
			return
		}

		if resUserID != user.ID {
			jsonifyErrf(w, http.StatusUnauthorized, "wrong user")
			return
		}

		log.Printf("unreserving item %d from user %s", req.ItemID, user.EMail)
		if err := rh.db.Unreserve(user.ID, req.ItemID); err != nil {
			jsonifyErrf(w, http.StatusInternalServerError, "failed to unreserve: %v", err)
			return
		}

		if err := rh.pollMgr.Publish("list-change", "unreserve"); err != nil {
			log.Printf("failed to publish event: %v", err)
		}
	}

	jsonifyErrf(w, http.StatusOK, "OK")
}

func (rh *ReserveHandler) NeedsAuthentication() bool {
	return true
}
