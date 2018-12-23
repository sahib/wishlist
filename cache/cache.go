package cache

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/dgraph-io/badger"
)

const (
	expireTime = 48 * time.Hour
)

type SessionCache struct {
	db *badger.DB
}

type Session struct {
	UserID      int64
	IsConfirmed bool
}

func NewSessionCache(path string) (*SessionCache, error) {
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &SessionCache{db: db}, nil
}

func (sc *SessionCache) Close() error {
	return sc.db.Close()
}

func (sc *SessionCache) Remember(userID int64, sessionID string) error {
	return sc.db.Update(func(txn *badger.Txn) error {
		sess := Session{
			UserID:      userID,
			IsConfirmed: false,
		}

		buf := &bytes.Buffer{}
		if err := gob.NewEncoder(buf).Encode(sess); err != nil {
			return err
		}

		return txn.SetWithTTL([]byte(sessionID), buf.Bytes(), expireTime)
	})
}

func (sc *SessionCache) Confirm(sessionID string) (int64, error) {
	userID := int64(-1)
	return userID, sc.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(sessionID))
		if err != nil {
			return err
		}

		data, err := item.Value()
		if err != nil {
			return err
		}

		sess := Session{}
		if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&sess); err != nil {
			return err
		}

		sess.IsConfirmed = true
		buf := &bytes.Buffer{}
		if err := gob.NewEncoder(buf).Encode(sess); err != nil {
			return err
		}

		userID = sess.UserID
		return txn.SetWithTTL([]byte(sessionID), buf.Bytes(), expireTime)
	})
}

func (sc *SessionCache) Session(sessionID string) (*Session, error) {
	sess := Session{}
	return &sess, sc.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(sessionID))
		if err != nil {
			return err
		}

		data, err := item.Value()
		if err != nil {
			return err
		}

		if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&sess); err != nil {
			return err
		}

		return nil
	})
}

func (sc *SessionCache) Forget(sessionID string) error {
	return sc.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(sessionID))
	})
}
