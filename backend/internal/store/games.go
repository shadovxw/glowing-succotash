package store

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shadovxw/monopoly/internal/game"
	"gorm.io/gorm"
)

type Game struct {
	ID           uuid.UUID       `gorm:"column:id;type:uuid;primaryKey"`
	RoomID       string          `gorm:"column:room_id"`
	Status       string          `gorm:"column:status"`
	StartedAt    *time.Time      `gorm:"column:started_at"`
	EndedAt      *time.Time      `gorm:"column:ended_at"`
	WinnerID     *string         `gorm:"column:winner_id"`
	PlayerIDs    []string        `gorm:"column:player_ids;type:text[]"`
	CurrentState json.RawMessage `gorm:"column:current_state;type:jsonb"`
	CreatedAt    time.Time       `gorm:"column:created_at"`
}

func (Game) TableName() string { return "games" }

type GameEvent struct {
	ID        int64           `gorm:"column:id;primaryKey"`
	GameID    uuid.UUID       `gorm:"column:game_id;type:uuid"`
	Seq       int             `gorm:"column:seq"`
	ActorID   *string         `gorm:"column:actor_id"`
	Type      string          `gorm:"column:type"`
	Payload   json.RawMessage `gorm:"column:payload;type:jsonb"`
	CreatedAt time.Time       `gorm:"column:created_at"`
}

func (GameEvent) TableName() string { return "game_events" }

type GameSeat struct {
	GameID        uuid.UUID `gorm:"column:game_id;type:uuid"`
	SeatIndex     int       `gorm:"column:seat_index"`
	BastionUserID string    `gorm:"column:bastion_user_id"`
}

func (GameSeat) TableName() string { return "game_seats" }

// CreateGame inserts a new game record.
func CreateGame(db *gorm.DB, state *game.GameState, playerIDs []string) error {
	now := time.Now()
	stateJSON, _ := json.Marshal(state)
	g := Game{
		ID:           state.ID,
		RoomID:       state.RoomID,
		Status:       state.Status,
		StartedAt:    &now,
		PlayerIDs:    playerIDs,
		CurrentState: stateJSON,
	}
	return db.Create(&g).Error
}

// SaveState persists the current GameState (called after every action).
func SaveState(db *gorm.DB, state *game.GameState) error {
	stateJSON, _ := json.Marshal(state)
	return db.Model(&Game{}).
		Where("id = ?", state.ID).
		Updates(map[string]any{
			"status":        state.Status,
			"current_state": stateJSON,
		}).Error
}

// EndGame marks a game as ended and records the winner.
func EndGame(db *gorm.DB, gameID uuid.UUID, winnerID string) error {
	now := time.Now()
	return db.Model(&Game{}).
		Where("id = ?", gameID).
		Updates(map[string]any{
			"status":    "ended",
			"ended_at":  now,
			"winner_id": winnerID,
		}).Error
}

// AppendEvent inserts a single game event into the log.
func AppendEvent(db *gorm.DB, gameID uuid.UUID, actorID *string, ev game.Event) error {
	return db.Create(&GameEvent{
		GameID:  gameID,
		ActorID: actorID,
		Type:    ev.Type,
		Payload: ev.Payload,
	}).Error
}

// LoadActiveGame returns the current GameState for a room, if one exists.
func LoadActiveGame(db *gorm.DB, roomID string) (*game.GameState, error) {
	var g Game
	err := db.Where("room_id = ? AND status IN ('active','paused')", roomID).
		Order("created_at DESC").
		First(&g).Error
	if err != nil {
		return nil, err
	}
	var state game.GameState
	if err := json.Unmarshal(g.CurrentState, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// UpsertSeats writes the seat assignments for a game.
func UpsertSeats(db *gorm.DB, gameID uuid.UUID, players []game.Player) error {
	for _, p := range players {
		seat := GameSeat{
			GameID:        gameID,
			SeatIndex:     p.SeatIndex,
			BastionUserID: p.BastionID,
		}
		if err := db.Save(&seat).Error; err != nil {
			return err
		}
	}
	return nil
}
