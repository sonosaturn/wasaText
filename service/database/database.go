package database

import (
	"database/sql"
	"errors"
)

// User is the user struct
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	PhotoURL string `json:"photo_url"`
}

// Conversation is the full struct we send to the frontend
type Conversation struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	PhotoURL            string `json:"photo_url"`
	IsGroup             bool   `json:"is_group"`
	LastMessageAt       string `json:"last_message_at"`
	LastMessagePreview  string `json:"last_message_preview"`
	OtherUserID         string `json:"other_user_id,omitempty"`
	LastMessageSenderID string `json:"last_message_sender_id"`
	LastMessageStatus   int    `json:"last_message_status"`
	UnreadCount         int    `json:"unread_count"`
}

// Reaction represents a single reaction to a message
type Reaction struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Emoji    string `json:"emoji"`
}

// Message is the struct for messages
type Message struct {
	ID             string     `json:"id"`
	ConversationID string     `json:"conversationId"`
	SenderID       string     `json:"senderId"`
	SenderUsername string     `json:"sender_username"` // <--- NEW: sender name
	Content        string     `json:"content"`
	PhotoURL       string     `json:"photo_url"`
	Timestamp      string     `json:"timestamp"`
	ReplyToID      *string    `json:"reply_to_id,omitempty"`
	Status         int        `json:"status"`
	Reactions      []Reaction `json:"reactions"`
	Forwarded      bool       `json:"forwarded"`
}

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	DoLogin(username string) (string, error)
	GetUser(id string) (User, error)
	SearchUsers(query string) ([]User, error)
	SetUserPhoto(id string, photoURL string) error
	SetUsername(id string, newName string) (bool, error)

	GetMyConversations(userID string) ([]Conversation, error)
	CreateGroup(name string, memberIDs []string) (string, error)
	CreateDirectConversation(myUserID string, otherUserID string) (string, error)
	GetConversationMembers(conversationID string) ([]User, error)

	SetConversationName(conversationID string, newName string) error
	SetConversationPhoto(conversationID string, photoURL string) error

	AddToConversation(conversationID string, userID string) error
	RemoveFromConversation(conversationID string, userID string) error
	LeaveConversation(conversationID string, userID string) error
	MarkConversationAsRead(conversationID string, userID string) error
	MarkConversationAsReceived(conversationID string, userID string) error

	SendMessage(conversationID string, senderID string, content string, photoURL string, replyToID string, forwarded bool) (*Message, error)
	GetConversationMessages(conversationID string, userID string) ([]Message, error)
	DeleteMessage(conversationID string, messageID string, userID string) error
	ReactToMessage(conversationID string, messageID string, userID string, emoji string) error
	UnreactToMessage(messageID string, userID string) error
	GetMessageForForwarding(msgID string, userID string) (string, string, error)

	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}

	// Schema unchanged (the new columns were added last round, these are just structs for JSON)
	var tableQueries = []string{
		`CREATE TABLE IF NOT EXISTS users (
            id TEXT NOT NULL PRIMARY KEY,
            username TEXT NOT NULL,
            photo_url TEXT
        );`,
		`CREATE TABLE IF NOT EXISTS conversations (
            id TEXT NOT NULL PRIMARY KEY,
            title TEXT,
            photo_url TEXT,
            is_group BOOLEAN NOT NULL DEFAULT 0,
            last_message_at DATETIME
        );`,
		`CREATE TABLE IF NOT EXISTS conversation_members (
            conversation_id TEXT NOT NULL,
            user_id TEXT NOT NULL,
            last_seen_at DATETIME,
            last_read_at DATETIME,
            PRIMARY KEY (conversation_id, user_id),
            FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );`,
		`CREATE TABLE IF NOT EXISTS messages (
            id TEXT NOT NULL PRIMARY KEY,
            conversation_id TEXT NOT NULL,
            sender_id TEXT NOT NULL,
            content TEXT,
            photo_url TEXT,
            reply_to_id TEXT,
            forwarded BOOLEAN NOT NULL DEFAULT 0,
            timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
            FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
        );`,
		`CREATE TABLE IF NOT EXISTS message_reactions (
            message_id TEXT NOT NULL,
            user_id TEXT NOT NULL,
            emoji TEXT NOT NULL,
            PRIMARY KEY (message_id, user_id),
            FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );`,
	}

	for _, query := range tableQueries {
		if _, err := db.Exec(query); err != nil {
			return nil, err
		}
	}

	return &appdbimpl{c: db}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
