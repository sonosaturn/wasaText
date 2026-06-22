package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
)

// DoLogin creates or finds a user with the given username and returns its identifier.
func (db *appdbimpl) DoLogin(username string) (string, error) {
	if len(username) < 3 || len(username) > 16 {
		return "", fmt.Errorf("username must be between 3 and 16 characters")
	}

	var id string
	err := db.c.QueryRow(`SELECT id FROM users WHERE username = ?;`, username).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("querying user: %w", err)
	}

	u, uerr := uuid.NewV4()
	if uerr != nil {
		return "", fmt.Errorf("generating user id: %w", uerr)
	}
	id = u.String()
	_, err = db.c.Exec(`INSERT INTO users (id, username) VALUES (?, ?);`, id, username)
	if err != nil {
		return "", fmt.Errorf("creating user: %w", err)
	}
	return id, nil
}

// GetUser returns the user identified by id. (RINOMINATA PER MATCHARE INTERFACCIA)
func (db *appdbimpl) GetUser(id string) (User, error) {
	var u User
	err := db.c.QueryRow(`
        SELECT id, username, COALESCE(photo_url, '') 
        FROM users 
        WHERE id = ?`, id).Scan(&u.ID, &u.Username, &u.PhotoURL)

	if err != nil {
		return User{}, err
	}
	return u, nil
}

// SearchUsers searches for users by username
func (db *appdbimpl) SearchUsers(query string) ([]User, error) {
	searchQuery := "%" + query + "%"
	rows, err := db.c.Query(`
        SELECT id, username, COALESCE(photo_url, '')
        FROM users 
        WHERE username LIKE ? 
        ORDER BY username ASC
    `, searchQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.PhotoURL); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// SetUserPhoto updates the photo URL for the user.
func (db *appdbimpl) SetUserPhoto(id string, photoURL string) error {
	_, err := db.c.Exec(`UPDATE users SET photo_url = ? WHERE id = ?`, photoURL, id)
	return err
}

// SetUsername updates the username if it's not already in use
func (db *appdbimpl) SetUsername(id string, newName string) (bool, error) {
	var otherId string
	err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", newName).Scan(&otherId)
	if err == nil {
		if otherId != id {
			return false, nil
		}
		return true, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	_, err = db.c.Exec("UPDATE users SET username = ? WHERE id = ?", newName, id)
	if err != nil {
		return false, err
	}
	return true, nil
}
