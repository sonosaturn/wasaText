package database

import (
	"database/sql"
	"errors"
	"time"
	"github.com/gofrs/uuid"
)

// SendMessage salva un nuovo messaggio (con supporto reply)
func (db *appdbimpl) SendMessage(conversationID string, senderID string, content string, photoURL string, replyToID string) (Message, error) {
	u, err := uuid.NewV4()
	if err != nil { return Message{}, err }
	msgID := u.String()
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Gestione NULL per replyToID
	var replyArg sql.NullString
	if replyToID != "" { 
		replyArg.String = replyToID; 
		replyArg.Valid = true 
	}

	_, err = db.c.Exec(`INSERT INTO messages (id, conversation_id, sender_id, content, photo_url, reply_to_id, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)`, 
		msgID, conversationID, senderID, content, photoURL, replyArg, timestamp)

	if err != nil { return Message{}, err }

	return Message{
		ID: msgID, ConversationID: conversationID, SenderID: senderID, Content: content, PhotoURL: photoURL, Timestamp: timestamp, Status: 1, ReplyToID: replyToID, Reactions: []Reaction{},
	}, nil
}

// MarkConversationAsRead aggiorna il timestamp di lettura dell'utente
func (db *appdbimpl) MarkConversationAsRead(conversationID string, userID string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := db.c.Exec(`
		UPDATE conversation_members 
		SET last_read_at = ? 
		WHERE conversation_id = ? AND user_id = ?
	`, now, conversationID, userID)
	return err
}

// ReactToMessage aggiunge o aggiorna una reazione (INSERT OR REPLACE è magia di SQLite)
func (db *appdbimpl) ReactToMessage(messageID string, userID string, emoji string) error {
	_, err := db.c.Exec(`INSERT OR REPLACE INTO message_reactions (message_id, user_id, emoji) VALUES (?, ?, ?)`, messageID, userID, emoji)
	return err
}

// UnreactToMessage rimuove la reazione
func (db *appdbimpl) UnreactToMessage(messageID string, userID string) error {
	_, err := db.c.Exec(`DELETE FROM message_reactions WHERE message_id = ? AND user_id = ?`, messageID, userID)
	return err
}

// GetConversationMessages recupera messaggi, status e reazioni
func (db *appdbimpl) GetConversationMessages(conversationID string, userID string) ([]Message, error) {
	// 1. Check membership
	var count int
	// FIX: Controlliamo l'errore di Scan
	if err := db.c.QueryRow(`SELECT COUNT(*) FROM conversation_members WHERE conversation_id = ? AND user_id = ?`, conversationID, userID).Scan(&count); err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("user not in conversation")
	}

	// 2. Min Read Time (per le spunte)
	var minReadTimeStr sql.NullString
	// FIX: Controlliamo l'errore di Scan. Se fallisce, assumiamo vuoto o gestiamo l'errore.
	if err := db.c.QueryRow(`SELECT MIN(last_read_at) FROM conversation_members WHERE conversation_id = ? AND user_id != ?`, conversationID, userID).Scan(&minReadTimeStr); err != nil {
		// Se non ci sono altri utenti o errore, ignoriamo e proseguiamo (non bloccante)
		minReadTimeStr.Valid = false
	}
	minReadTime := ""
	if minReadTimeStr.Valid {
		minReadTime = minReadTimeStr.String
	}

	// 3. Scarica i messaggi
	rows, err := db.c.Query(`SELECT id, conversation_id, sender_id, content, photo_url, reply_to_id, timestamp FROM messages WHERE conversation_id = ? ORDER BY timestamp ASC`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var photoUrl, replyTo sql.NullString

		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, &photoUrl, &replyTo, &msg.Timestamp); err != nil {
			return nil, err
		}

		if photoUrl.Valid {
			msg.PhotoURL = photoUrl.String
		}
		if replyTo.Valid {
			msg.ReplyToID = replyTo.String
		}

		// Calcolo Spunte
		msg.Status = 1
		if minReadTime != "" && msg.Timestamp <= minReadTime {
			msg.Status = 2
		}

		// 4. Scarica le Reazioni per QUESTO messaggio
		rRows, err := db.c.Query(`
			SELECT r.user_id, u.username, r.emoji 
			FROM message_reactions r 
			JOIN users u ON r.user_id = u.id 
			WHERE r.message_id = ?`, msg.ID)

		if err == nil {
			msg.Reactions = []Reaction{}
			for rRows.Next() {
				var r Reaction
				// FIX: Controlliamo errore Scan
				if err := rRows.Scan(&r.UserID, &r.Username, &r.Emoji); err != nil {
					continue // Saltiamo reazione corrotta
				}
				msg.Reactions = append(msg.Reactions, r)
			}
			rRows.Close()
			// FIX: rowserrcheck - controlliamo errore dopo il loop
			if err := rRows.Err(); err != nil {
				return nil, err
			}
		}

		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

func (db *appdbimpl) DeleteMessage(messageID string, requestorID string) error {
	res, err := db.c.Exec(`DELETE FROM messages WHERE id = ? AND sender_id = ?`, messageID, requestorID)
	if err != nil { return err }
	if aff, _ := res.RowsAffected(); aff == 0 { return errors.New("message not found or forbidden") }
	return nil
}