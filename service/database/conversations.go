package database

import (
	"errors"
	"github.com/gofrs/uuid"
)

// GetMyConversations retrieves the list of chats (both groups and private)
func (db *appdbimpl) GetMyConversations(userID string) ([]Conversation, error) {
	var conversations []Conversation

	query := `
	SELECT 
		c.id,
		c.is_group,
		COALESCE(CASE WHEN c.is_group = 1 THEN c.title ELSE u.username END, 'Unknown') as title,
		COALESCE(CASE WHEN c.is_group = 1 THEN c.photo_url ELSE u.photo_url END, '') as photo_url,
		COALESCE((SELECT content FROM messages m WHERE m.conversation_id = c.id ORDER BY m.timestamp DESC LIMIT 1), '') as last_msg,
		COALESCE((SELECT timestamp FROM messages m WHERE m.conversation_id = c.id ORDER BY m.timestamp DESC LIMIT 1), '') as last_time,
		COALESCE((SELECT sender_id FROM messages m WHERE m.conversation_id = c.id ORDER BY m.timestamp DESC LIMIT 1), '') as last_sender,
		
		-- Fetch the STATUS of the last message
		COALESCE((SELECT 
			CASE 
				WHEN m2.sender_id = ? THEN 
					CASE 
						WHEN (SELECT COUNT(*) FROM conversation_members cm WHERE cm.conversation_id = m2.conversation_id AND cm.last_read_at >= m2.timestamp AND cm.user_id != m2.sender_id) = (SELECT COUNT(*) FROM conversation_members cm WHERE cm.conversation_id = m2.conversation_id AND cm.user_id != m2.sender_id)
						THEN 2 
						ELSE 0 
					END
				ELSE 1 
			END
		FROM messages m2 WHERE m2.conversation_id = c.id ORDER BY m2.timestamp DESC LIMIT 1), 0) as last_status,

		COALESCE(u.id, '') as other_user_id,
		(SELECT COUNT(*) 
		 FROM messages m 
		 WHERE m.conversation_id = c.id 
		 AND m.timestamp > COALESCE(my_cm.last_read_at, '1970-01-01 00:00:00')
		) as unread_count

	FROM conversation_members my_cm
	JOIN conversations c ON my_cm.conversation_id = c.id
	
	LEFT JOIN conversation_members other_cm ON c.id = other_cm.conversation_id AND other_cm.user_id != my_cm.user_id AND c.is_group = 0
	LEFT JOIN users u ON other_cm.user_id = u.id

	WHERE my_cm.user_id = ?
	ORDER BY last_time DESC
	`

	rows, err := db.c.Query(query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c Conversation
		var lastMsg, lastTime, lastSender, otherUserID string
		var lastStatus int

		err = rows.Scan(
			&c.ID,
			&c.IsGroup,
			&c.Title,
			&c.PhotoURL,
			&lastMsg,
			&lastTime,
			&lastSender,
			&lastStatus,
			&otherUserID,
			&c.UnreadCount,
		)
		if err != nil {
			return nil, err
		}

		c.LastMessagePreview = lastMsg
		if c.LastMessagePreview == "" && lastTime != "" {
			c.LastMessagePreview = "[Photo]"
		}
		c.LastMessageAt = lastTime
		c.LastMessageSenderID = lastSender
		c.LastMessageStatus = lastStatus
		c.OtherUserID = otherUserID

		conversations = append(conversations, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (db *appdbimpl) CreateGroup(name string, memberIDs []string) (string, error) {
	groupID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	idStr := groupID.String()

	_, err = db.c.Exec("INSERT INTO conversations (id, title, is_group, last_message_at) VALUES (?, ?, 1, DATETIME('now'))", idStr, name)
	if err != nil {
		return "", err
	}

	queryMember := "INSERT INTO conversation_members (conversation_id, user_id, last_read_at) VALUES (?, ?, DATETIME('now'))"
	for _, memberID := range memberIDs {
		if _, err := db.c.Exec(queryMember, idStr, memberID); err != nil {
			continue
		}
	}

	return idStr, nil
}

func (db *appdbimpl) CreateDirectConversation(myUserID string, otherUserID string) (string, error) {
	var existingID string
	err := db.c.QueryRow(`
		SELECT c.id 
		FROM conversations c
		JOIN conversation_members cm1 ON c.id = cm1.conversation_id
		JOIN conversation_members cm2 ON c.id = cm2.conversation_id
		WHERE c.is_group = 0 
		AND cm1.user_id = ? 
		AND cm2.user_id = ? 
		LIMIT 1
	`, myUserID, otherUserID).Scan(&existingID)

	if err == nil {
		return existingID, nil
	}

	convID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	idStr := convID.String()

	_, err = db.c.Exec("INSERT INTO conversations (id, is_group, last_message_at) VALUES (?, 0, DATETIME('now'))", idStr)
	if err != nil {
		return "", err
	}

	_, err = db.c.Exec("INSERT INTO conversation_members (conversation_id, user_id, last_read_at) VALUES (?, ?, DATETIME('now'))", idStr, myUserID)
	if err != nil {
		return "", err
	}
	_, err = db.c.Exec("INSERT INTO conversation_members (conversation_id, user_id, last_read_at) VALUES (?, ?, DATETIME('now'))", idStr, otherUserID)
	if err != nil {
		return "", err
	}

	return idStr, nil
}

func (db *appdbimpl) GetConversationMembers(conversationID string) ([]User, error) {
	query := `
		SELECT u.id, u.username, COALESCE(u.photo_url, '') 
		FROM users u
		JOIN conversation_members cm ON u.id = cm.user_id
		WHERE cm.conversation_id = ?
		ORDER BY u.username ASC`

	rows, err := db.c.Query(query, conversationID)
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

func (db *appdbimpl) SetConversationName(conversationID string, newName string) error {
	_, err := db.c.Exec("UPDATE conversations SET title = ? WHERE id = ? AND is_group = 1", newName, conversationID)
	return err
}

func (db *appdbimpl) SetConversationPhoto(conversationID string, photoURL string) error {
	_, err := db.c.Exec("UPDATE conversations SET photo_url = ? WHERE id = ? AND is_group = 1", photoURL, conversationID)
	return err
}

func (db *appdbimpl) AddToConversation(conversationID string, userID string) error {
	var isGroup bool
	err := db.c.QueryRow("SELECT is_group FROM conversations WHERE id = ?", conversationID).Scan(&isGroup)
	if err != nil {
		return err
	}
	if !isGroup {
		return errors.New("cannot add members to a direct chat")
	}

	_, err = db.c.Exec("INSERT OR IGNORE INTO conversation_members (conversation_id, user_id, last_read_at) VALUES (?, ?, DATETIME('now'))", conversationID, userID)
	return err
}

// FIX: Changed to delete the entire chat if it's direct (1-to-1)
func (db *appdbimpl) RemoveFromConversation(conversationID string, userID string) error {
	// 1. Check if it's a group
	var isGroup bool
	err := db.c.QueryRow("SELECT is_group FROM conversations WHERE id = ?", conversationID).Scan(&isGroup)
	if err != nil {
		return err
	}

	if !isGroup {
		// IF IT'S A DIRECT CHAT: Delete the entire conversation.
		// Thanks to CASCADE, the messages and the other member will also be deleted.
		_, err := db.c.Exec("DELETE FROM conversations WHERE id = ?", conversationID)
		return err
	} else {
		// IF IT'S A GROUP: Remove only the member (standard behavior)
		_, err := db.c.Exec("DELETE FROM conversation_members WHERE conversation_id = ? AND user_id = ?", conversationID, userID)
		return err
	}
}

func (db *appdbimpl) LeaveConversation(conversationID string, userID string) error {
	return db.RemoveFromConversation(conversationID, userID)
}

func (db *appdbimpl) MarkConversationAsRead(conversationID string, userID string) error {
	_, err := db.c.Exec(`
        UPDATE conversation_members
        SET last_read_at = CURRENT_TIMESTAMP
        WHERE conversation_id = ? AND user_id = ?
    `, conversationID, userID)
	return err
}
