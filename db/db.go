package db

import (
	"database/sql"
	"sync"

	// Load the sqlite3 module, we don't need anything from it.
	_ "github.com/mattn/go-sqlite3"
)

const (
	createSQL = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users(
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS items(
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    link TEXT NOT NULL,
	created_by INTEGER,

    FOREIGN KEY(created_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS reservations(
    user_id INTEGER,
    item_id INTEGER,

    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(item_id) REFERENCES items(id)
);
`
)

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	EMail string `json:"email"`
}

type Item struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Link  string `json:"link,omitempty"`
	IsOwn bool   `json:"is_own"`
}

type Database struct {
	mu sync.Mutex

	db             *sql.DB
	userInsertStmt *sql.Stmt
	itemInsertStmt *sql.Stmt
	rsrvInsertStmt *sql.Stmt
	rsrvDeleteStmt *sql.Stmt
	itemDeleteStmt *sql.Stmt
}

func NewDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(createSQL); err != nil {
		return nil, err
	}

	userInsertStmt, err := db.Prepare("INSERT INTO users(name, email) VALUES(?, ?);")
	if err != nil {
		return nil, err
	}

	itemInsertStmt, err := db.Prepare("INSERT INTO items(name, link, created_by) VALUES(?, ?, ?);")
	if err != nil {
		return nil, err
	}

	rsrvInsertStmt, err := db.Prepare("INSERT INTO reservations(user_id, item_id) VALUES(?, ?);")
	if err != nil {
		return nil, err
	}

	rsrvDeleteStmt, err := db.Prepare("DELETE FROM reservations WHERE user_id = ? AND item_id = ?;")
	if err != nil {
		return nil, err
	}

	itemDeleteStmt, err := db.Prepare("DELETE FROM items WHERE id = ? AND created_by = ?;")
	if err != nil {
		return nil, err
	}

	return &Database{
		db:             db,
		userInsertStmt: userInsertStmt,
		itemInsertStmt: itemInsertStmt,
		rsrvInsertStmt: rsrvInsertStmt,
		rsrvDeleteStmt: rsrvDeleteStmt,
		itemDeleteStmt: itemDeleteStmt,
	}, nil
}

func (db *Database) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.db.Close()
}

func (db *Database) AddUser(name, email string) (int64, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	res, err := db.userInsertStmt.Exec(name, email)
	if err != nil {
		return -1, err
	}

	return res.LastInsertId()
}

func (db *Database) GetUserByEMail(email string) (*User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	u := &User{}

	row := db.db.QueryRow("SELECT id, name, email FROM users WHERE email = ?;", email)
	if err := row.Scan(&u.ID, &u.Name, &u.EMail); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return u, nil
}
func (db *Database) GetUserByID(ID int64) (*User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	u := &User{}

	row := db.db.QueryRow("SELECT id, name, email FROM users WHERE id = ?;", ID)
	if err := row.Scan(&u.ID, &u.Name, &u.EMail); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return u, nil
}

func (db *Database) AddItem(name, link string, createdBy int64) (int64, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	res, err := db.itemInsertStmt.Exec(name, link, createdBy)
	if err != nil {
		return -1, err
	}

	return res.LastInsertId()
}

func (db *Database) DeleteItem(userID, itemID int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.itemDeleteStmt.Exec(itemID, userID)
	return err
}

func (db *Database) GetItems(userID int64) ([]*Item, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	rows, err := db.db.Query("SELECT id, name, link, created_by FROM items;")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := []*Item{}
	for rows.Next() {
		item := &Item{}
		createdBy := int64(0)
		if err := rows.Scan(&item.ID, &item.Name, &item.Link, &createdBy); err != nil {
			return nil, err
		}

		item.IsOwn = userID == createdBy
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (db *Database) Reserve(userID, itemID int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.rsrvInsertStmt.Exec(userID, itemID)
	return err
}

func (db *Database) Unreserve(userID, itemID int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.rsrvDeleteStmt.Exec(userID, itemID)
	return err
}

func (db *Database) GetUserForReservation(itemID int64) (int64, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	row := db.db.QueryRow("SELECT user_id FROM reservations WHERE item_id = ?;", itemID)
	userID := int64(0)
	err := row.Scan(&userID)
	if err == sql.ErrNoRows {
		return -1, nil
	}

	if err != nil {
		return -1, err
	}

	return userID, nil
}
