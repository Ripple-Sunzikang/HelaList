package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ChatRepository struct {
	db *sql.DB
}

type ChatSession struct {
	ID        int       `json:"id"`
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id,omitempty"`
	Title     string    `json:"title,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatMessage struct {
	ID        int                    `json:"id"`
	SessionID string                 `json:"session_id"`
	Role      string                 `json:"role"` // "user" or "assistant"
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) CreateSession(userID, title string) (*ChatSession, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	session := &ChatSession{
		SessionID: sessionID,
		UserID:    userID,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := r.db.QueryRow(`
		INSERT INTO chat_sessions (session_id, user_id, title, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, sessionID, userID, title, now, now).Scan(&session.ID)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *ChatRepository) GetSession(sessionID string) (*ChatSession, error) {
	var session ChatSession
	err := r.db.QueryRow(`
		SELECT id, session_id, COALESCE(user_id, ''), COALESCE(title, ''), created_at, updated_at
		FROM chat_sessions
		WHERE session_id = $1
	`, sessionID).Scan(
		&session.ID,
		&session.SessionID,
		&session.UserID,
		&session.Title,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *ChatRepository) UpdateSessionTitle(sessionID, title string) error {
	_, err := r.db.Exec(`
		UPDATE chat_sessions 
		SET title = $1, updated_at = $2
		WHERE session_id = $3
	`, title, time.Now(), sessionID)
	return err
}

func (r *ChatRepository) AddMessage(sessionID, role, content string, metadata map[string]interface{}) (*ChatMessage, error) {
	var metadataJSON []byte
	var err error

	if metadata != nil {
		metadataJSON, err = json.Marshal(metadata)
		if err != nil {
			return nil, err
		}
	} else {
		metadataJSON = []byte("{}")
	}

	message := &ChatMessage{
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	err = r.db.QueryRow(`
		INSERT INTO chat_messages (session_id, role, content, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, sessionID, role, content, metadataJSON, message.CreatedAt).Scan(&message.ID)

	if err != nil {
		return nil, err
	}

	return message, nil
}

func (r *ChatRepository) GetMessages(sessionID string, limit int) ([]ChatMessage, error) {
	if limit <= 0 {
		limit = 50 // 默认限制
	}

	rows, err := r.db.Query(`
		SELECT id, session_id, role, content, metadata, created_at
		FROM chat_messages
		WHERE session_id = $1
		ORDER BY created_at ASC
		LIMIT $2
	`, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var message ChatMessage
		var metadataJSON []byte

		err := rows.Scan(
			&message.ID,
			&message.SessionID,
			&message.Role,
			&message.Content,
			&metadataJSON,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析metadata
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &message.Metadata)
			if err != nil {
				message.Metadata = make(map[string]interface{})
			}
		}

		messages = append(messages, message)
	}

	return messages, rows.Err()
}

func (r *ChatRepository) GetRecentMessages(sessionID string, count int) ([]ChatMessage, error) {
	if count <= 0 {
		count = 10 // 默认获取最近10条
	}

	rows, err := r.db.Query(`
		SELECT id, session_id, role, content, metadata, created_at
		FROM chat_messages
		WHERE session_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, sessionID, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var message ChatMessage
		var metadataJSON []byte

		err := rows.Scan(
			&message.ID,
			&message.SessionID,
			&message.Role,
			&message.Content,
			&metadataJSON,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析metadata
		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &message.Metadata)
			if err != nil {
				message.Metadata = make(map[string]interface{})
			}
		}

		messages = append(messages, message)
	}

	// 因为我们是按时间倒序查询的，需要反转顺序
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, rows.Err()
}

func (r *ChatRepository) DeleteSession(sessionID string) error {
	_, err := r.db.Exec("DELETE FROM chat_sessions WHERE session_id = $1", sessionID)
	return err
}

func (r *ChatRepository) GetUserSessions(userID string, limit int) ([]ChatSession, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.Query(`
		SELECT id, session_id, COALESCE(user_id, ''), COALESCE(title, ''), created_at, updated_at
		FROM chat_sessions
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []ChatSession
	for rows.Next() {
		var session ChatSession
		err := rows.Scan(
			&session.ID,
			&session.SessionID,
			&session.UserID,
			&session.Title,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}
