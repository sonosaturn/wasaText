package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gofrs/uuid"
)

func (db *appdbimpl) SendMessage(conversationID string, senderID string, content string, photoURL string, replyToID string, forwarded bool) (*Message, error) {
	msgUUID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	msgID := msgUUID.String()

	var replyTo sql.NullString
	if replyToID != "" {
		replyTo.String = replyToID
		replyTo.Valid = true
	}

	query := `INSERT INTO messages (id, conversation_id, sender_id, content, photo_url, reply_to_id, forwarded, timestamp) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, DATETIME('now'))`

	_, err = db.c.Exec(query, msgID, conversationID, senderID, content, photoURL, replyTo, forwarded)
	if err != nil {
		return nil, err
	}

	_, _ = db.c.Exec("UPDATE conversations SET last_message_at = DATETIME('now') WHERE id = ?", conversationID)

	return &Message{
		ID:             msgID,
		ConversationID: conversationID,
		SenderID:       senderID,
		// SenderUsername: NOT set here, the API wrapper sets it because the DB doesn't know it immediately without another query
		Content:   content,
		PhotoURL:  photoURL,
		ReplyToID: &replyToID,
		Forwarded: forwarded,
		Timestamp: time.Now().Format(time.RFC3339),
		Status:    0,
	}, nil
}

func (db *appdbimpl) GetConversationMessages(conversationID string, userID string) ([]Message, error) {
	var messages []Message

	// FIX: Join with USERS table to get the sender's username
	query := `
		SELECT 
			m.id, 
			m.sender_id,
			u.username, -- FIX: Seleziona username
			m.content, 
			COALESCE(m.photo_url, '') as photo_url, 
			m.reply_to_id, 
			m.forwarded,
			m.timestamp,
			CASE 
				WHEN m.sender_id = ? THEN 
					CASE 
						WHEN (SELECT COUNT(*) FROM conversation_members cm WHERE cm.conversation_id = m.conversation_id AND cm.last_read_at >= m.timestamp AND cm.user_id != m.sender_id) = (SELECT COUNT(*) FROM conversation_members cm WHERE cm.conversation_id = m.conversation_id AND cm.user_id != m.sender_id)
						THEN 2 
						ELSE 0 
					END
				ELSE 1 
			END as status
		FROM messages m
		JOIN users u ON m.sender_id = u.id -- FIX: JOIN
		WHERE m.conversation_id = ?
		ORDER BY m.timestamp ASC
	`

	rows, err := db.c.Query(query, userID, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m Message
		var replyTo sql.NullString

		// FIX: Scan anche m.SenderUsername
		err = rows.Scan(&m.ID, &m.SenderID, &m.SenderUsername, &m.Content, &m.PhotoURL, &replyTo, &m.Forwarded, &m.Timestamp, &m.Status)
		if err != nil {
			return nil, err
		}

		if replyTo.Valid {
			str := replyTo.String
			m.ReplyToID = &str
		}

		m.ConversationID = conversationID
		reactions, _ := db.getMessageReactions(m.ID)
		m.Reactions = reactions

		messages = append(messages, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

// ... Restanti funzioni (Delete, React, etc.) invariate ...
func (db *appdbimpl) DeleteMessage(conversationID string, messageID string, userID string) error {
	res, err := db.c.Exec("DELETE FROM messages WHERE id = ? AND sender_id = ? AND conversation_id = ?", messageID, userID, conversationID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("message not found or not owned by user")
	}
	return nil
}

func (db *appdbimpl) ReactToMessage(conversationID string, messageID string, userID string, emoji string) error {
	_, err := db.c.Exec(`INSERT INTO message_reactions (message_id, user_id, emoji) VALUES (?, ?, ?) ON CONFLICT(message_id, user_id) DO UPDATE SET emoji = excluded.emoji`, messageID, userID, emoji)
	return err
}

func (db *appdbimpl) UnreactToMessage(messageID string, userID string) error {
	_, err := db.c.Exec("DELETE FROM message_reactions WHERE message_id = ? AND user_id = ?", messageID, userID)
	return err
}

func (db *appdbimpl) getMessageReactions(messageID string) ([]Reaction, error) {
	query := `SELECT r.user_id, u.username, r.emoji FROM message_reactions r JOIN users u ON r.user_id = u.id WHERE r.message_id = ?`
	rows, err := db.c.Query(query, messageID)
	if err != nil {
		return []Reaction{}, err
	}
	defer rows.Close()
	var reactions []Reaction
	for rows.Next() {
		var r Reaction
		if err := rows.Scan(&r.UserID, &r.Username, &r.Emoji); err == nil {
			reactions = append(reactions, r)
		}
	}
	return reactions, rows.Err()
}

func (db *appdbimpl) GetMessageForForwarding(msgID string, userID string) (string, string, error) {
	var content, photoURL string
	query := `SELECT m.content, COALESCE(m.photo_url, '') FROM messages m JOIN conversation_members cm ON m.conversation_id = cm.conversation_id WHERE m.id = ? AND cm.user_id = ?`
	err := db.c.QueryRow(query, msgID, userID).Scan(&content, &photoURL)
	return content, photoURL, err
}
