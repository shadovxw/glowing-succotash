package store

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Player struct {
	BastionUserID string    `gorm:"column:bastion_user_id;primaryKey"`
	DisplayName   string    `gorm:"column:display_name"`
	Avatar        string    `gorm:"column:avatar"`
	GamesPlayed   int       `gorm:"column:games_played"`
	GamesWon      int       `gorm:"column:games_won"`
	LastSeen      time.Time `gorm:"column:last_seen"`
}

func (Player) TableName() string { return "players" }

// Upsert inserts or updates the player record on every WS connect.
func UpsertPlayer(db *gorm.DB, id, displayName, avatar string) error {
	p := Player{
		BastionUserID: id,
		DisplayName:   displayName,
		Avatar:        avatar,
		LastSeen:      time.Now(),
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "bastion_user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"display_name", "avatar", "last_seen"}),
	}).Create(&p).Error
}

func IncrementGamesPlayed(db *gorm.DB, id string) error {
	return db.Model(&Player{}).
		Where("bastion_user_id = ?", id).
		UpdateColumn("games_played", gorm.Expr("games_played + 1")).Error
}

func IncrementGamesWon(db *gorm.DB, id string) error {
	return db.Model(&Player{}).
		Where("bastion_user_id = ?", id).
		UpdateColumn("games_won", gorm.Expr("games_won + 1")).Error
}

func GetPlayerStats(db *gorm.DB, id string) (*Player, error) {
	var p Player
	if err := db.Where("bastion_user_id = ?", id).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
