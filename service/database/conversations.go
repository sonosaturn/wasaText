package database

import (
	"errors"
	"github.com/gofrs/uuid"
)

// GetMyConversations recupera la lista delle chat (sia gruppi che private)
func (db *appdbimpl) GetMyConversations(userID string) ([]Conversation, error) {
	var conversations []Conversation

	// Questa query recupera le conversazioni unendo le tabelle.
	// Se è una chat privata (is_group = 0), recupera nome e foto dell'altro utente.
	// Se è un gruppo, usa titolo e foto del gruppo.
	// Calcola i messaggi non letti confrontando il timestamp dei messaggi con l'ultimo accesso dell'utente (last_read_at).

	query := `
	SELECT 
		c.id,
		c.is_group,
		-- Titolo: Se gruppo usa c.title, se privato usa u.username (dell'altro utente)
		COALESCE(CASE WHEN c.is_group = 1 THEN c.title ELSE u.username END, 'Sconosciuto') as title,
		-- Foto: Se gruppo usa c.photo_url, se privato usa u.photo_url
		COALESCE(CASE WHEN c.is_group = 1 THEN c.photo_url ELSE u.photo_url END, '') as photo_url,
		-- Anteprima ultimo messaggio (subquery per semplicità, ottimizzabile)
		COALESCE((SELECT content FROM messages m WHERE m.conversation_id = c.id ORDER BY m.timestamp DESC LIMIT 1), '') as last_msg,
		COALESCE((SELECT timestamp FROM messages m WHERE m.conversation_id = c.id ORDER BY m.timestamp DESC LIMIT 1), '') as last_time,
		-- ID dell'ultimo mittente (per mostrare "Tu:")
		COALESCE((SELECT sender_id FROM messages m WHERE m.conversation_id = c.id ORDER BY m.timestamp DESC LIMIT 1), '') as last_sender,
		-- ID dell'altro utente (per chat private)
		COALESCE(u.id, '') as other_user_id,
		-- Conteggio non letti: messaggi successivi al mio last_read_at
		(SELECT COUNT(*) 
		 FROM messages m 
		 WHERE m.conversation_id = c.id 
		 AND m.timestamp > COALESCE(my_cm.last_read_at, '1970-01-01 00:00:00')
		) as unread_count

	FROM conversation_members my_cm
	JOIN conversations c ON my_cm.conversation_id = c.id
	
	-- Join per trovare l'altro utente nelle chat private
	LEFT JOIN conversation_members other_cm ON c.id = other_cm.conversation_id AND other_cm.user_id != my_cm.user_id AND c.is_group = 0
	LEFT JOIN users u ON other_cm.user_id = u.id

	WHERE my_cm.user_id = ?
	ORDER BY last_time DESC
	`

	rows, err := db.c.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c Conversation
		var lastMsg, lastTime, lastSender, otherUserID string

		err = rows.Scan(
			&c.ID,
			&c.IsGroup,
			&c.Title,
			&c.PhotoURL,
			&lastMsg,
			&lastTime,
			&lastSender,
			&otherUserID,
			&c.UnreadCount,
		)
		if err != nil {
			return nil, err
		}

		c.LastMessagePreview = lastMsg
		if c.LastMessagePreview == "" && lastTime != "" {
			c.LastMessagePreview = "[Foto]" // Se c'è timestamp ma non testo, è una foto
		}
		c.LastMessageAt = lastTime
		c.LastMessageSenderID = lastSender
		c.OtherUserID = otherUserID

		conversations = append(conversations, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

// CreateGroup crea una nuova conversazione di gruppo
func (db *appdbimpl) CreateGroup(name string, memberIDs []string) (string, error) {
	// 1. Genera UUID
	groupID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	idStr := groupID.String()

	// 2. Crea Conversazione
	_, err = db.c.Exec("INSERT INTO conversations (id, title, is_group, last_message_at) VALUES (?, ?, 1, DATETIME('now'))", idStr, name)
	if err != nil {
		return "", err
	}

	// 3. Aggiungi Membri
	queryMember := "INSERT INTO conversation_members (conversation_id, user_id, last_read_at) VALUES (?, ?, DATETIME('now'))"
	for _, memberID := range memberIDs {
		if _, err := db.c.Exec(queryMember, idStr, memberID); err != nil {
			// Continua anche se uno fallisce, o gestisci rollback (qui semplificato)
			continue
		}
	}

	return idStr, nil
}

// CreateDirectConversation crea o recupera una chat 1-a-1
func (db *appdbimpl) CreateDirectConversation(myUserID string, otherUserID string) (string, error) {
	// 1. Controlla se esiste già
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

	// 2. Se non esiste, crea nuova
	convID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	idStr := convID.String()

	// Inserisci conversazione (senza titolo/foto, sono dinamici per le chat dirette)
	_, err = db.c.Exec("INSERT INTO conversations (id, is_group, last_message_at) VALUES (?, 0, DATETIME('now'))", idStr)
	if err != nil {
		return "", err
	}

	// Inserisci i due membri
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

// GetConversationMembers recupera la lista degli utenti in una chat
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

// SetConversationName cambia il titolo (solo per gruppi)
func (db *appdbimpl) SetConversationName(conversationID string, newName string) error {
	_, err := db.c.Exec("UPDATE conversations SET title = ? WHERE id = ? AND is_group = 1", newName, conversationID)
	return err
}

// SetConversationPhoto cambia la foto (solo per gruppi)
func (db *appdbimpl) SetConversationPhoto(conversationID string, photoURL string) error {
	_, err := db.c.Exec("UPDATE conversations SET photo_url = ? WHERE id = ? AND is_group = 1", photoURL, conversationID)
	return err
}

// AddToConversation aggiunge un utente a un gruppo
func (db *appdbimpl) AddToConversation(conversationID string, userID string) error {
	// Verifica che sia un gruppo
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

// RemoveFromConversation rimuove un utente
func (db *appdbimpl) RemoveFromConversation(conversationID string, userID string) error {
	_, err := db.c.Exec("DELETE FROM conversation_members WHERE conversation_id = ? AND user_id = ?", conversationID, userID)
	return err
}

// LeaveConversation è alias di Remove
func (db *appdbimpl) LeaveConversation(conversationID string, userID string) error {
	return db.RemoveFromConversation(conversationID, userID)
}

// MarkConversationAsRead aggiorna il timestamp di lettura dell'utente
func (db *appdbimpl) MarkConversationAsRead(conversationID string, userID string) error {
	// Usiamo CURRENT_TIMESTAMP per coerenza con il resto del database.
	// Rimuoviamo last_seen_at che causava l'errore di sintassi.
	_, err := db.c.Exec(`
        UPDATE conversation_members 
        SET last_read_at = CURRENT_TIMESTAMP
        WHERE conversation_id = ? AND user_id = ?
    `, conversationID, userID)
	return err
}
