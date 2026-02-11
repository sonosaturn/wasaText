package database

import (
	"database/sql"
	"errors"
)

// User rappresenta il modello utente
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	PhotoURL string `json:"photo_url"`
}

// ConversationSummary per la lista chat
type ConversationSummary struct {
	ID          string `json:"id"`
	IsGroup     bool   `json:"is_group"`
	Title       string `json:"title"`
	PhotoURL    string `json:"photo_url"`
	OtherUserID string `json:"other_user_id,omitempty"`
}

// Reaction rappresenta una singola reazione a un messaggio
type Reaction struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Emoji    string `json:"emoji"`
}

// Message aggiornato con ReplyTo e Reactions
type Message struct {
	ID             string     `json:"id"`
	ConversationID string     `json:"conversationId"`
	SenderID       string     `json:"senderId"`
	Content        string     `json:"content"`
	PhotoURL       string     `json:"photo_url"`
	Timestamp      string     `json:"timestamp"`
	Status         int        `json:"status"` // 1=Inviato, 2=Letto
	
	// NUOVI CAMPI AGGIUNTI
	ReplyToID string     `json:"reply_to_id"` // ID del messaggio citato (opzionale)
	Reactions []Reaction `json:"reactions"`   // Lista reazioni
}

// AppDatabase interfaccia principale
type AppDatabase interface {
	Ping() error

	// Users
	DoLogin(username string) (string, error)
	GetUserById(id string) (User, error)
	SearchUsers(query string) ([]User, error)
	SetUserPhoto(id string, photoURL string) error

	// Groups
	CreateGroup(name string, members []string) (string, error)
	AddMemberToGroup(conversationID string, userID string) error
	RemoveMemberFromGroup(conversationID string, userID string) error
	SetGroupName(conversationID string, newName string) error
	SetGroupPhoto(conversationID string, photoURL string) error
	GetGroupMembers(conversationID string) ([]User, error)

	// Conversations & Messages
	ListConversations(userID string) ([]ConversationSummary, error)
	CreateDirectConversation(myUserID string, otherUserID string) (string, error)
	
	// SendMessage AGGIORNATO: accetta replyToID
	SendMessage(conversationID string, senderID string, content string, photoURL string, replyToID string) (Message, error)
	GetConversationMessages(conversationID string, userID string) ([]Message, error)
	DeleteMessage(messageID string, requestorID string) error
	MarkConversationAsRead(conversationID string, userID string) error
	
	// NUOVI METODI PER REAZIONI
	ReactToMessage(messageID string, userID string, emoji string) error
	UnreactToMessage(messageID string, userID string) error

	GetName() (string, error)
	SetName(name string) error
}

type appdbimpl struct {
	c *sql.DB
}

var _ AppDatabase = (*appdbimpl)(nil)

func New(db *sql.DB) (AppDatabase, error) {
	if db == nil { return nil, errors.New("database is required") }
	
	// Abilita Foreign Keys
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil { return nil, err }

	// 1. Users
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, username TEXT NOT NULL UNIQUE, photo_url TEXT);`); err != nil { return nil, err }

	// 2. Conversations
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS conversations (id TEXT PRIMARY KEY, is_group INTEGER DEFAULT 0, title TEXT, photo_url TEXT);`); err != nil { return nil, err }

	// 3. Members
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS conversation_members (conversation_id TEXT, user_id TEXT, last_read_at DATETIME DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (conversation_id, user_id), FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE, FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE);`); err != nil { return nil, err }

	// 4. Messages (AGGIUNTO reply_to_id)
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY, 
		conversation_id TEXT NOT NULL, 
		sender_id TEXT NOT NULL, 
		content TEXT, 
		photo_url TEXT, 
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		reply_to_id TEXT,  -- <--- NUOVA COLONNA
		FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
		FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (reply_to_id) REFERENCES messages(id) ON DELETE SET NULL
	);`); err != nil { return nil, err }

	// 5. Message Reactions (NUOVA TABELLA)
	// La chiave primaria composta (message_id, user_id) assicura che un utente possa mettere solo 1 reazione per messaggio
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS message_reactions (
		message_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		emoji TEXT NOT NULL,
		PRIMARY KEY (message_id, user_id),
		FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`); err != nil { return nil, err }

	return &appdbimpl{c: db}, nil
}

func (db *appdbimpl) Ping() error { return db.c.Ping() }
func (db *appdbimpl) GetName() (string, error) { return "WASAText DB", nil }
func (db *appdbimpl) SetName(name string) error { return nil }