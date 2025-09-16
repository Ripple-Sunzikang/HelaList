package model

import "time"

// ChatSession 对话会话
type ChatSession struct {
	ID        int64         `json:"id" gorm:"primaryKey"`
	SessionID string        `json:"session_id" gorm:"uniqueIndex;size:100"`
	UserID    string        `json:"user_id" gorm:"size:100"`
	Title     string        `json:"title" gorm:"size:200"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Messages  []ChatMessage `json:"messages,omitempty" gorm:"foreignKey:SessionID;references:SessionID"`
}

// ChatMessage 对话消息
type ChatMessage struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	SessionID string    `json:"session_id" gorm:"index;size:100"`
	Role      string    `json:"role" gorm:"size:20"` // user, assistant, system
	Content   string    `json:"content" gorm:"type:text"`
	Metadata  *string   `json:"metadata,omitempty" gorm:"type:text"` // JSON格式的额外元数据，使用指针表示可空
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

// TableName 设置表名
func (ChatSession) TableName() string {
	return "chat_sessions"
}

// TableName 设置表名
func (ChatMessage) TableName() string {
	return "chat_messages"
}

// CreateChatSessionRequest 创建会话请求
type CreateChatSessionRequest struct {
	UserID string `json:"user_id"`
	Title  string `json:"title"`
}

// ChatRequest 聊天请求（包含会话ID）
type ChatRequest struct {
	SessionID string `json:"session_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Message   string `json:"message"`
	UseRAG    bool   `json:"use_rag,omitempty"` // 是否启用RAG
}

// ChatResponse 聊天响应
type ChatResponse struct {
	SessionID string       `json:"session_id"`
	Message   string       `json:"message"`
	Context   []RAGContext `json:"context,omitempty"` // RAG上下文
}

// RAGContext RAG上下文信息
type RAGContext struct {
	Source     string  `json:"source"`     // 文档来源
	Content    string  `json:"content"`    // 相关内容
	Similarity float64 `json:"similarity"` // 相似度分数
}
