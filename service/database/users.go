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

// GetUserByID returns the user identified by id.
func (db *appdbimpl) GetUserById(id string) (User, error) {
    var u User
    // COALESCE gestisce i casi senza foto
    err := db.c.QueryRow(`
        SELECT id, username, COALESCE(photo_url, '') 
        FROM users 
        WHERE id = ?`, id).Scan(&u.ID, &u.Username, &u.PhotoURL)
    
    if err != nil {
        return User{}, err
    }
    return u, nil
}

// SearchUsers searches for users by username (case-insensitive substring match).
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
	if err != nil {
		return err
	}
	return nil
}