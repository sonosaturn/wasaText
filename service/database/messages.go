package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gofrs/uuid"
)

func (db *appdbimpl) SendMessage(conversationID string, senderID string, content string, photoURL string, replyToID string) (*Message, error) {
	// 1. Genera ID Messaggio
	msgUUID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	msgID := msgUUID.String()

	// 2. Gestione ReplyTo (null se vuoto)
	var replyTo sql.NullString
	if replyToID != "" {
		replyTo.String = replyToID
		replyTo.Valid = true
	}

	// 3. Inserisci nel DB
	query := `INSERT INTO messages (id, conversation_id, sender_id, content, photo_url, reply_to_id, timestamp) 
			  VALUES (?, ?, ?, ?, ?, ?, DATETIME('now'))`
	
	_, err = db.c.Exec(query, msgID, conversationID, senderID, content, photoURL, replyTo)
	if err != nil {
		return nil, err
	}

	// 4. Aggiorna l'orario della conversazione (per l'ordinamento chat list)
	_, _ = db.c.Exec("UPDATE conversations SET last_message_at = DATETIME('now') WHERE id = ?", conversationID)

	// 5. Costruiamo l'oggetto da restituire
	return &Message{
		ID:             msgID,
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		PhotoURL:       photoURL,
		ReplyToID:      &replyToID,
		Timestamp:      time.Now().Format(time.RFC3339),
		Status:         0, // Inviato
	}, nil
}

func (db *appdbimpl) GetConversationMessages(conversationID string, userID string) ([]Message, error) {
	var messages []Message

	query := `
		SELECT 
			m.id, 
			m.sender_id, 
			m.content, 
			COALESCE(m.photo_url, '') as photo_url, 
			m.reply_to_id, 
			m.timestamp,
			-- Calcolo dello stato (semplificato)
			CASE 
				WHEN m.sender_id = ? THEN 
					CASE 
						WHEN (SELECT COUNT(*) 
                              FROM conversation_members cm 
                              WHERE cm.conversation_id = m.conversation_id 
                              AND cm.last_read_at >= m.timestamp 
                              AND cm.user_id != m.sender_id) 
                              = 
                             (SELECT COUNT(*) 
                              FROM conversation_members cm 
                              WHERE cm.conversation_id = m.conversation_id 
                              AND cm.user_id != m.sender_id)
						THEN 2 -- Letto
						ELSE 0 -- Inviato
					END
				ELSE 1 -- Ricevuto (se non sono io il mittente)
			END as status
		FROM messages m
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

		err = rows.Scan(&m.ID, &m.SenderID, &m.Content, &m.PhotoURL, &replyTo, &m.Timestamp, &m.Status)
		if err != nil {
			return nil, err
		}

		if replyTo.Valid {
			str := replyTo.String
			m.ReplyToID = &str
		}
		
		m.ConversationID = conversationID
		
		// Carica Reazioni
		reactions, _ := db.getMessageReactions(m.ID)
		m.Reactions = reactions

		messages = append(messages, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

func (db *appdbimpl) DeleteMessage(conversationID string, messageID string, userID string) error {
	// Permetti cancellazione solo se l'utente è il mittente
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
	_, err := db.c.Exec(`
		INSERT INTO message_reactions (message_id, user_id, emoji) 
		VALUES (?, ?, ?) 
		ON CONFLICT(message_id, user_id) DO UPDATE SET emoji = excluded.emoji`,
		messageID, userID, emoji)
	return err
}

func (db *appdbimpl) UnreactToMessage(messageID string, userID string) error {
	_, err := db.c.Exec("DELETE FROM message_reactions WHERE message_id = ? AND user_id = ?", messageID, userID)
	return err
}

func (db *appdbimpl) getMessageReactions(messageID string) ([]Reaction, error) {
	query := `
		SELECT r.user_id, u.username, r.emoji 
		FROM message_reactions r
		JOIN users u ON r.user_id = u.id
		WHERE r.message_id = ?`

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
	// Verifica che l'utente faccia parte della conversazione in cui si trova il messaggio originale
	query := `
		SELECT m.content, COALESCE(m.photo_url, '')
		FROM messages m
		JOIN conversation_members cm ON m.conversation_id = cm.conversation_id
		WHERE m.id = ? AND cm.user_id = ?`
	
	err := db.c.QueryRow(query, msgID, userID).Scan(&content, &photoURL)
	return content, photoURL, err
}