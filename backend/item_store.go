package main

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// reminders: Scan() closes connections so you dont need to call Close()

// returned by CreateUser when another user already has the email
var ErrEmailTaken = errors.New("a user with that email already exists")

// User model
type User struct {
	UserID *int64 `json:"user_id"`
	Email  string `json:"email"`
}

// UserSortPreferences model
type UserSortPreferences struct {
	UserID    *int64 `json:"user_id"`
	SortOrder string `json:"sort_order"`
}

// hidden devices model
type HiddenDevice struct {
	UserID   *int64 `json:"user_id"`
	DeviceID string `json:"device_id"`
	Ignore   bool   `json:"ignore"`
}

// device nickname model
type DeviceNickname struct {
	UserID      *int64 `json:"user_id"`
	DeviceID    string `json:"device_id"`
	DisplayName string `json:"display_name"`
	Ignore      bool   `json:"ignore"`
}

// device marker model
type DeviceMarker struct {
	UserID      *int64 `json:"user_id"`
	DeviceID    string `json:"device_id"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
	Ignore      bool   `json:"ignore"`
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
	conn.SetMaxOpenConns(1)
	return ItemStore{conn}, err
}

// runs creation of tables if they do not exist initially
// separate tables for each feature to follow normalzation rules
func (s *ItemStore) CreateDB() error {

	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL)`); err != nil {
		log.Fatal(err)
	}

	if _, err := s.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique
		ON users (email COLLATE NOCASE)`); err != nil {
		log.Fatal(err)
	}
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS user_sort_preferences (
		user_id INTEGER PRIMARY KEY,
		sort_order TEXT NOT NULL
		)`); err != nil {
		log.Fatal(err)
	}
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS user_hidden_devices (
		user_id INTEGER NOT NULL,
		device_id TEXT NOT NULL,
		ignore BOOLEAN DEFAULT FALSE,
		PRIMARY KEY (user_id, device_id)
		)`); err != nil {
		log.Fatal(err)
	}
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS user_device_nicknames (
		user_id INTEGER NOT NULL,
		device_id TEXT NOT NULL,
		display_name TEXT NOT NULL,
		ignore BOOLEAN DEFAULT FALSE,
		PRIMARY KEY (user_id, device_id)
	)`); err != nil {
		log.Fatal(err)
	}
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS user_device_markers (
		user_id INTEGER NOT NULL,
		device_id TEXT NOT NULL,
		content_type TEXT NOT NULL,
		data BLOB NOT NULL,
		ignore BOOLEAN DEFAULT FALSE,
		PRIMARY KEY (user_id, device_id)
	)`); err != nil {
		log.Fatal(err)
	}
	return nil
}

// returns one user
func (s *ItemStore) GetUser(ctx context.Context, user_id int64) (*User, error) {
	var userID int64
	var email string
	err := s.db.QueryRowContext(ctx, `SELECT user_id, email FROM users WHERE user_id = ?`, user_id).Scan(&userID, &email)
	if err != nil {
		return nil, err
	}
	return &User{
		UserID: &userID,
		Email:  email,
	}, nil
}

// creates a new user in db
func (s *ItemStore) CreateUser(ctx context.Context, u *User) error {
	var userID int64
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO users (email)
			VALUES (?)
			RETURNING user_id`,
		u.Email).Scan(&userID)
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
		return ErrEmailTaken
	}
	if err != nil {
		return err
	}
	u.UserID = &userID
	return nil
}

// remove this after updating main.go
// returns a set of all users and emails
func (s *ItemStore) GetAllUser(ctx context.Context) ([]*User, error) {
	var users []*User
	rows, err := s.db.QueryContext(ctx, `SELECT user_id, email FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID int64
		var email string
		if err := rows.Scan(&userID, &email); err == nil {
			users = append(users, &User{
				UserID: &userID,
				Email:  email,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// returns sort preference of user
func (s *ItemStore) GetSortPreferences(ctx context.Context, user_id int64) (string, error) {
	var sortOrder string
	err := s.db.QueryRowContext(ctx, `SELECT sort_order FROM user_sort_preferences WHERE user_id = ?`, user_id).Scan(&sortOrder)

	if err != nil {
		return "desc", err
	}
	return sortOrder, nil
}

// inserts if UserSortPreferences struct has no ID otherwise update based on ID
func (s *ItemStore) UpsertSortPreference(ctx context.Context, p *UserSortPreferences) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO user_sort_preferences (user_id, sort_order)
		VALUES (?, ?) 
		ON CONFLICT (user_id)
		DO UPDATE SET
		sort_order = excluded.sort_order
		RETURNING user_id, sort_order`,
		p.UserID, p.SortOrder)

	return err
}

func (s *ItemStore) GetHiddenDeviceIDs(ctx context.Context, user_id int64) ([]string, error) {
	var hidden_device_ids []string
	rows, err := s.db.QueryContext(ctx,
		`SELECT device_id 
		FROM user_hidden_devices 
		WHERE user_id = ?
		AND ignore = FALSE`, user_id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err == nil {
			hidden_device_ids = append(hidden_device_ids, s)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hidden_device_ids, nil
}

func (s *ItemStore) UpsertHiddenDevice(ctx context.Context, d *HiddenDevice) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO user_hidden_devices (user_id, device_id, ignore)
		VALUES (?, ?, ?) 
		ON CONFLICT (user_id, device_id)
		DO UPDATE SET
		ignore = excluded.ignore`,
		d.UserID, d.DeviceID, d.Ignore)

	return err

}

func (s *ItemStore) GetDeviceNicknames(ctx context.Context, user_id int64) (map[string]string, error) {
	var nickname_map = make(map[string]string)
	rows, err := s.db.QueryContext(ctx,
		`SELECT device_id, display_name
		FROM user_device_nicknames
		WHERE user_id = ?
		AND ignore = FALSE`, user_id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var d string
		var n string
		if err := rows.Scan(&d, &n); err == nil {
			nickname_map[d] = n
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nickname_map, nil
}

func (s *ItemStore) UpsertDeviceNickname(ctx context.Context, n *DeviceNickname) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO user_device_nicknames (user_id, device_id, display_name, ignore)
		VALUES (?, ?, ?, ?) 
		ON CONFLICT (user_id, device_id)
		DO UPDATE SET
		display_name = excluded.display_name,
		ignore = excluded.ignore`,
		n.UserID, n.DeviceID, n.DisplayName, n.Ignore)
	return err
}

func (s *ItemStore) GetDeviceMarker(ctx context.Context, user_id int64, device_id string) (*DeviceMarker, error) {
	var userID int64
	var deviceID string
	var contentType string
	var data []byte
	var ignore bool
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, device_id, content_type, data, ignore
		FROM user_device_markers
		WHERE user_id = ?
		AND device_id = ?
		AND ignore = FALSE
		`, user_id, device_id).Scan(&userID, &deviceID, &contentType, &data, &ignore)

	if err != nil {
		return nil, err
	}
	return &DeviceMarker{
		UserID:      &userID,
		DeviceID:    deviceID,
		ContentType: contentType,
		Data:        data,
		Ignore:      ignore,
	}, nil
}

func (s *ItemStore) getDeviceWithMarkerIDs(ctx context.Context, user_id int64) ([]string, error) {
	var device_ids []string
	rows, err := s.db.QueryContext(ctx,
		`SELECT device_id 
		FROM user_device_markers
		WHERE user_id = ?
		AND ignore = FALSE`, user_id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err == nil {
			device_ids = append(device_ids, s)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return device_ids, nil
}

func (s *ItemStore) UpsertDeviceMarker(ctx context.Context, m *DeviceMarker) error {
	_, err := s.db.ExecContext(ctx, `
	INSERT INTO user_device_markers (user_id, device_id, content_type, data, ignore)
		VALUES (?, ?, ?, ?, ?) 
		ON CONFLICT (user_id, device_id)
		DO UPDATE SET
		content_type = excluded.content_type,
		data = excluded.data,
		ignore = excluded.ignore`,
		m.UserID, m.DeviceID, m.ContentType, m.Data, m.Ignore)
	return err
}

// soft removes a marker by setting ignore, returns sql.ErrNoRows if there was no marker
func (s *ItemStore) IgnoreDeviceMarker(ctx context.Context, user_id int64, device_id string) error {
	res, err := s.db.ExecContext(ctx, `
	UPDATE user_device_markers
		SET ignore = TRUE
		WHERE user_id = ?
		AND device_id = ?`,
		user_id, device_id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
