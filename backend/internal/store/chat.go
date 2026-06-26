package store

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatMessage struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	RoomID      string     `gorm:"column:room_id"`
	GameID      *uuid.UUID `gorm:"column:game_id;type:uuid"`
	SenderID    *string    `gorm:"column:sender_id"`
	DisplayName string     `gorm:"column:display_name"`
	Message     string     `gorm:"column:message"`
	IsSystem    bool       `gorm:"column:is_system"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
}

func (ChatMessage) TableName() string { return "chat_messages" }

func InsertChat(db *gorm.DB, roomID string, gameID *uuid.UUID, senderID *string, displayName, message string, isSystem bool) error {
	return db.Create(&ChatMessage{
		RoomID:      roomID,
		GameID:      gameID,
		SenderID:    senderID,
		DisplayName: displayName,
		Message:     message,
		IsSystem:    isSystem,
	}).Error
}

// RecentMessages returns the last N messages for a room (newest last).
func RecentMessages(db *gorm.DB, roomID string, limit int) ([]ChatMessage, error) {
	var msgs []ChatMessage
	err := db.Where("room_id = ?", roomID).
		Order("created_at DESC").
		Limit(limit).
		Find(&msgs).Error
	if err != nil {
		return nil, err
	}
	// Reverse so messages are in chronological order
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}
