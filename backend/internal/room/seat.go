package room

import (
	fws "github.com/gofiber/websocket/v2"
)

// Seat represents one player's slot in a room.
type Seat struct {
	Index       int
	BastionID   string
	DisplayName string
	Avatar      string
	Conn        *fws.Conn // nil when disconnected
	Send        chan []byte
}

func NewSeat(index int, bastionID, displayName, avatar string) *Seat {
	return &Seat{
		Index:       index,
		BastionID:   bastionID,
		DisplayName: displayName,
		Avatar:      avatar,
		Send:        make(chan []byte, 64),
	}
}

func (s *Seat) Connected() bool {
	return s.Conn != nil
}

// WritePump drains the send channel and forwards to the WebSocket connection.
func (s *Seat) WritePump() {
	defer func() {
		if s.Conn != nil {
			s.Conn.Close()
		}
	}()
	for msg := range s.Send {
		if s.Conn == nil {
			continue
		}
		if err := s.Conn.WriteMessage(1, msg); err != nil {
			return
		}
	}
}
