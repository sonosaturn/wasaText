package database

import (
	"database/sql"
)

// ListConversations restituisce tutte le conversazioni di un utente
func (db *appdbimpl) ListConversations(userID string) ([]ConversationSummary, error) {
	// Query: Prendi ID, IsGroup, Title, PhotoURL dalle conversazioni di cui faccio parte
	rows, err := db.c.Query(`
		SELECT c.id, c.is_group, c.title, c.photo_url
		FROM conversations c
		JOIN conversation_members cm ON c.id = cm.conversation_id
		WHERE cm.user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []ConversationSummary

	for rows.Next() {
		var c ConversationSummary
		var title sql.NullString
		var photoUrl sql.NullString

		if err := rows.Scan(&c.ID, &c.IsGroup, &title, &photoUrl); err != nil {
			return nil, err
		}

		if title.Valid {
			c.Title = title.String
		}
		if photoUrl.Valid {
			c.PhotoURL = photoUrl.String
		}

		// Se NON è un gruppo, dobbiamo trovare chi è l'altro utente per mostrare il suo nome/foto
		if !c.IsGroup {
			var otherID string
			err := db.c.QueryRow(`
				SELECT user_id FROM conversation_members 
				WHERE conversation_id = ? AND user_id != ? 
				LIMIT 1`, c.ID, userID).Scan(&otherID)
			
			if err == nil {
				c.OtherUserID = otherID
			}
		}

		conversations = append(conversations, c)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

// CreateDirectConversation crea una chat 1-a-1 se non esiste già
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
		return existingID, nil // Esiste già
	}
    
    // Se non esiste, usiamo CreateGroup (che gestisce UUID e inserimenti)
    // Passando nome vuoto "" viene creato con is_group=0 (vedi logica in CreateGroup)
    return db.CreateGroup("", []string{myUserID, otherUserID})
}