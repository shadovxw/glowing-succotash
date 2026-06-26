package ws

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	fws "github.com/gofiber/websocket/v2"
	"github.com/shadovxw/monopoly/internal/auth"
	"github.com/shadovxw/monopoly/internal/proto"
	"github.com/shadovxw/monopoly/internal/room"
	"github.com/shadovxw/monopoly/internal/store"
	"gorm.io/gorm"
)

type Handler struct {
	manager   *room.Manager
	validator *auth.Validator
	db        *gorm.DB
}

func NewHandler(manager *room.Manager, validator *auth.Validator, db *gorm.DB) *Handler {
	return &Handler{manager: manager, validator: validator, db: db}
}

// Upgrade is the Fiber middleware that upgrades HTTP → WebSocket.
func (h *Handler) Upgrade(c *fiber.Ctx) error {
	if fws.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// Connect is the actual WebSocket handler after upgrade.
func (h *Handler) Connect(c *fws.Conn) {
	cookie := c.Cookies("auth_session")
	identity, err := h.validator.Validate(cookie)
	if err != nil {
		_ = c.WriteMessage(fws.CloseMessage, fws.FormatCloseMessage(4001, "unauthorized"))
		return
	}

	_ = store.UpsertPlayer(h.db, identity.Sub, identity.DisplayName, identity.Avatar)

	roomCode := c.Params("roomCode")
	var r *room.Room
	if roomCode == "new" {
		r, err = h.manager.Create()
		if err != nil {
			log.Println("create room:", err)
			return
		}
	} else {
		r, err = h.manager.Get(roomCode)
		if err != nil {
			_ = c.WriteMessage(fws.CloseMessage, fws.FormatCloseMessage(4004, "room not found"))
			return
		}
	}

	seat, err := r.JoinSync(identity.Sub, identity.DisplayName, identity.Avatar)
	if err != nil {
		_ = c.WriteMessage(fws.CloseMessage, fws.FormatCloseMessage(4003, err.Error()))
		return
	}
	seat.Conn = c

	// Replay recent chat history
	msgs, _ := store.RecentMessages(h.db, r.ID, 50)
	for _, m := range msgs {
		data, _ := proto.Encode(proto.MsgChatMessage, map[string]any{
			"senderId":    m.SenderID,
			"displayName": m.DisplayName,
			"message":     m.Message,
			"isSystem":    m.IsSystem,
			"createdAt":   m.CreatedAt,
		})
		seat.Send <- data
	}

	// Announce join
	joinMsg, _ := proto.Encode(proto.MsgPlayerJoined, map[string]any{
		"bastionId":   identity.Sub,
		"displayName": identity.DisplayName,
		"avatar":      identity.Avatar,
		"seatIndex":   seat.Index,
		"roomCode":    r.ID,
	})
	r.Broadcast(joinMsg)

	// Send current game state if a game is in progress
	if r.State != nil {
		stateMsg, _ := proto.Encode(proto.MsgGameState, r.State)
		seat.Send <- stateMsg
	}

	go seat.WritePump()

	// Read loop
	for {
		_, raw, err := c.ReadMessage()
		if err != nil {
			break
		}
		var env proto.Envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			continue
		}
		r.Dispatch(identity.Sub, env)
	}

	// Disconnected
	seat.Conn = nil
	leaveMsg, _ := proto.Encode(proto.MsgPlayerLeft, map[string]any{
		"bastionId": identity.Sub,
	})
	r.Broadcast(leaveMsg)
}
