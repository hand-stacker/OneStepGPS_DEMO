package main

import (
	"context"
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// reminders: Scan() closes connections so you dont need to call Close()

// UserPreference model
type UserPreferences struct {
	UserID    *int64 `json:"user_id"`
	SortOrder string `json:"sort_order"`
}

// blocked devices model
type BlockedDevices struct {
	UserID   int64  `json:"user_id"`
	DeviceID string `json:"device_id"`
}

// struct that handles all db functions
type ItemStore struct {
	db *sql.DB
}

// makes an ItemStore with sql connextion
func NewItemStore() (ItemStore, error) {
	dbPath := getenv("DB_PATH", "data.db")
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)

	}

	return ItemStore{conn}, err
}

// runs creation of tables if they do not exist initially
func (s *ItemStore) CreateDB() error {
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS user_preferences (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
		sort_order TEXT NOT NULL)`); err != nil {
		log.Fatal(err)
	}
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS blocked_devices (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id TEXT NOT NULL)`); err != nil {
		log.Fatal(err)
	}
	return nil
}

// returns a UserPreference if exists
func (s *ItemStore) GetPreference(ctx context.Context, user_id int64) (*UserPreferences, error) {
	var userID int64
	var sortOrder string

	err := s.db.QueryRowContext(ctx, `SELECT user_id, sort_order FROM user_preferences WHERE user_id = ?`, user_id).Scan(&userID, &sortOrder)

	if err != nil {
		return nil, err
	}

	prefs := &UserPreferences{
		UserID:    &userID,
		SortOrder: sortOrder,
	}
	return prefs, nil
}

// returns a set of all UserPreferences(for testing rn, will REMOVE later)
func (s *ItemStore) GetAllPreferences(ctx context.Context) ([]*UserPreferences, error) {
	var prefs []*UserPreferences
	rows, err := s.db.QueryContext(ctx, `SELECT user_id, sort_order FROM user_preferences`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID int64
		var sortOrder string
		if err := rows.Scan(&userID, &sortOrder); err == nil {
			prefs = append(prefs, &UserPreferences{
				UserID:    &userID,
				SortOrder: sortOrder,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return prefs, nil
}

// inserts if UserPreferences struct has no ID otherwise update based on ID
func (s *ItemStore) UpsertPreference(ctx context.Context, p *UserPreferences) error {
	var ret *sql.Row
	if p.UserID == nil {
		// here I need to set defaults if p hase trivial values
		ret = s.db.QueryRowContext(ctx,
			`INSERT INTO user_preferences (sort_order)
			VALUES (?)
			RETURNING user_id, sort_order`,
			p.SortOrder)
	} else {
		// here i need to only update the nontrivial values in p
		ret = s.db.QueryRowContext(ctx,
			`UPDATE user_preferences
			SET sort_order = ?
			WHERE user_id = ?
			RETURNING user_id, sort_order`,
			p.SortOrder, *p.UserID)
	}

	var userID int64
	var sortOrder string
	if err := ret.Scan(&userID, &sortOrder); err != nil {
		return err
	}

	p.UserID = &userID
	p.SortOrder = sortOrder
	return nil
}
