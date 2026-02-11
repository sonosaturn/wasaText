package database

import (
	"database/sql"
	"fmt"
	"github.com/gofrs/uuid"
)

// CreateGroup crea un nuovo gruppo
func (db *appdbimpl) CreateGroup(name string, members []string) (string, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	groupID := u.String()

	// Se name è vuoto, è una chat diretta (is_group = 0), altrimenti è un gruppo (is_group = 1)
	isGroup := 0
	if name != "" {
		isGroup = 1
	}

	// Inseriamo photo_url vuoto di default
	_, err = db.c.Exec(`INSERT INTO conversations (id, is_group, title, photo_url) VALUES (?, ?, ?, ?)`, 
		groupID, isGroup, name, "") 
	if err != nil {
		return "", fmt.Errorf("creating group: %w", err)
	}

	// Aggiungi membri
	for _, memberID := range members {
		if err := db.AddMemberToGroup(groupID, memberID); err != nil {
			return "", err
		}
	}

	return groupID, nil
}

func (db *appdbimpl) AddMemberToGroup(conversationID string, userID string) error {
	_, err := db.c.Exec(`INSERT OR IGNORE INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`, conversationID, userID)
	return err
}

func (db *appdbimpl) RemoveMemberFromGroup(conversationID string, userID string) error {
	_, err := db.c.Exec(`DELETE FROM conversation_members WHERE conversation_id = ? AND user_id = ?`, conversationID, userID)
	return err
}

func (db *appdbimpl) GetGroupMembers(conversationID string) ([]User, error) {
	rows, err := db.c.Query(`
		SELECT u.id, u.username, u.photo_url 
		FROM users u
		JOIN conversation_members cm ON u.id = cm.user_id
		WHERE cm.conversation_id = ?
	`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []User
	for rows.Next() {
		var u User
		var photoUrl sql.NullString // Serve database/sql qui
		if err := rows.Scan(&u.ID, &u.Username, &photoUrl); err != nil {
			return nil, err
		}
		if photoUrl.Valid {
			u.PhotoURL = photoUrl.String
		}
		members = append(members, u)
	}
	return members, rows.Err()
}

// SetGroupName cambia il titolo del gruppo
func (db *appdbimpl) SetGroupName(conversationID string, newName string) error {
	_, err := db.c.Exec(`UPDATE conversations SET title = ? WHERE id = ?`, newName, conversationID)
	return err
}

// SetGroupPhoto cambia la foto del gruppo
func (db *appdbimpl) SetGroupPhoto(conversationID string, photoURL string) error {
	_, err := db.c.Exec(`UPDATE conversations SET photo_url = ? WHERE id = ?`, photoURL, conversationID)
	return err
}