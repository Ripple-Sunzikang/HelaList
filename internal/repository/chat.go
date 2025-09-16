package repository

import (
	"fmt"
	"time"

	"HelaList/internal/model"

	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// CreateSession 创建新的对话会话
func (r *ChatRepository) CreateSession(session *model.ChatSession) error {
	// 如果没有提供SessionID，生成一个唯一的ID
	if session.SessionID == "" {
		session.SessionID = fmt.Sprintf("session_%d_%d", time.Now().Unix(), time.Now().UnixNano())
	}

	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()

	return r.db.Create(session).Error
}

// GetSession 根据SessionID获取会话
func (r *ChatRepository) GetSession(sessionID string) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessionWithMessages 获取会话及其消息
func (r *ChatRepository) GetSessionWithMessages(sessionID string, limit int) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}

	// 获取最近的消息（按时间倒序，然后限制数量）
	var messages []model.ChatMessage
	query := r.db.Where("session_id = ?", sessionID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	err = query.Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// 反转消息顺序，使最老的消息在前面
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	session.Messages = messages
	return &session, nil
}

// GetUserSessions 获取用户的所有会话
func (r *ChatRepository) GetUserSessions(userID string, limit, offset int) ([]model.ChatSession, error) {
	var sessions []model.ChatSession
	query := r.db.Where("user_id = ?", userID).Order("updated_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&sessions).Error
	return sessions, err
}

// UpdateSession 更新会话信息
func (r *ChatRepository) UpdateSession(sessionID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.Model(&model.ChatSession{}).Where("session_id = ?", sessionID).Updates(updates).Error
}

// DeleteSession 删除会话及其所有消息
func (r *ChatRepository) DeleteSession(sessionID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先删除所有消息
		if err := tx.Where("session_id = ?", sessionID).Delete(&model.ChatMessage{}).Error; err != nil {
			return err
		}
		// 再删除会话
		return tx.Where("session_id = ?", sessionID).Delete(&model.ChatSession{}).Error
	})
}

// AddMessage 添加消息到会话
func (r *ChatRepository) AddMessage(message *model.ChatMessage) error {
	message.CreatedAt = time.Now()

	return r.db.Transaction(func(tx *gorm.DB) error {
		// 添加消息
		if err := tx.Create(message).Error; err != nil {
			return err
		}

		// 更新会话的最后更新时间
		return tx.Model(&model.ChatSession{}).
			Where("session_id = ?", message.SessionID).
			Update("updated_at", time.Now()).Error
	})
}

// GetMessages 获取会话的消息列表
func (r *ChatRepository) GetMessages(sessionID string, limit int) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	query := r.db.Where("session_id = ?", sessionID).Order("created_at ASC")

	if limit > 0 {
		// 如果有限制，获取最近的N条消息
		var totalCount int64
		r.db.Model(&model.ChatMessage{}).Where("session_id = ?", sessionID).Count(&totalCount)

		if int64(limit) < totalCount {
			offset := int(totalCount) - limit
			query = query.Offset(offset)
		}
	}

	err := query.Find(&messages).Error
	return messages, err
}

// GetRecentMessages 获取最近的消息（用于构建上下文）
func (r *ChatRepository) GetRecentMessages(sessionID string, count int) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	err := r.db.Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(count).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	// 反转顺序，使最老的消息在前面
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// CleanupOldSessions 清理旧的会话（可选功能）
func (r *ChatRepository) CleanupOldSessions(daysOld int) error {
	cutoffTime := time.Now().AddDate(0, 0, -daysOld)

	return r.db.Transaction(func(tx *gorm.DB) error {
		// 查找要删除的会话
		var sessionIDs []string
		err := tx.Model(&model.ChatSession{}).
			Where("updated_at < ?", cutoffTime).
			Pluck("session_id", &sessionIDs).Error
		if err != nil {
			return err
		}

		if len(sessionIDs) == 0 {
			return nil
		}

		// 删除消息
		if err := tx.Where("session_id IN ?", sessionIDs).Delete(&model.ChatMessage{}).Error; err != nil {
			return err
		}

		// 删除会话
		return tx.Where("session_id IN ?", sessionIDs).Delete(&model.ChatSession{}).Error
	})
}
