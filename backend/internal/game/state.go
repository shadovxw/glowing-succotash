package game

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// NewGame builds a fresh GameState for a set of seated players.
// players must be ordered by seat index (seat 0 goes first).
func NewGame(roomID string, players []Player, rules *RuleSet) *GameState {
	tiles := make([]TileState, 40)
	for i := range tiles {
		tiles[i] = TileState{Position: i}
	}

	chanceDeck := shuffledDeck(len(rules.Chance))
	commDeck := shuffledDeck(len(rules.Community))

	return &GameState{
		ID:          uuid.New(),
		RoomID:      roomID,
		Status:      StatusActive,
		Players:     players,
		Tiles:       tiles,
		ActiveIdx:   0,
		Phase:       PhaseRoll,
		HousesLeft:  rules.Config.HouseLimit,
		HotelsLeft:  rules.Config.HotelLimit,
		ChanceDeck:  chanceDeck,
		CommDeck:    commDeck,
		PendingTrades: []Trade{},
		EventSeq:    0,
		StartedAt:   time.Now(),
	}
}

// ActivePlayer returns a pointer to the player whose turn it currently is.
func (s *GameState) ActivePlayer() *Player {
	return &s.Players[s.ActiveIdx]
}

// PlayerByBastionID finds a player by their Bastion identity. Returns nil if not found.
func (s *GameState) PlayerByBastionID(id string) *Player {
	for i := range s.Players {
		if s.Players[i].BastionID == id {
			return &s.Players[i]
		}
	}
	return nil
}

// TileAt returns a pointer to the TileState at the given position.
func (s *GameState) TileAt(pos int) *TileState {
	return &s.Tiles[pos]
}

// OwnerOf returns the player who owns the tile at pos, or nil if unowned.
func (s *GameState) OwnerOf(pos int) *Player {
	ts := s.TileAt(pos)
	if ts.OwnerID == nil {
		return nil
	}
	for i := range s.Players {
		if s.Players[i].ID == *ts.OwnerID {
			return &s.Players[i]
		}
	}
	return nil
}

// AdvanceTurn moves to the next active (non-bankrupt) player.
func (s *GameState) AdvanceTurn() {
	count := len(s.Players)
	for i := 1; i <= count; i++ {
		next := (s.ActiveIdx + i) % count
		if !s.Players[next].Bankrupt {
			s.ActiveIdx = next
			s.Phase = PhaseRoll
			s.DoublesCount = 0
			s.TurnNumber++
			return
		}
	}
}

// ActivePlayerCount returns the number of players still in the game.
func (s *GameState) ActivePlayerCount() int {
	n := 0
	for _, p := range s.Players {
		if !p.Bankrupt {
			n++
		}
	}
	return n
}

// DrawChance draws the next Chance card, reshuffling the deck if exhausted.
func (s *GameState) DrawChance(rules *RuleSet) Card {
	if s.ChanceIdx >= len(s.ChanceDeck) {
		s.ChanceDeck = shuffledDeck(len(rules.Chance))
		s.ChanceIdx = 0
	}
	card := rules.Chance[s.ChanceDeck[s.ChanceIdx]]
	s.ChanceIdx++
	return card
}

// DrawCommunity draws the next Community Chest card.
func (s *GameState) DrawCommunity(rules *RuleSet) Card {
	if s.CommIdx >= len(s.CommDeck) {
		s.CommDeck = shuffledDeck(len(rules.Community))
		s.CommIdx = 0
	}
	card := rules.Community[s.CommDeck[s.CommIdx]]
	s.CommIdx++
	return card
}

func shuffledDeck(n int) []int {
	deck := make([]int, n)
	for i := range deck {
		deck[i] = i
	}
	rand.Shuffle(n, func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

// NewEvent constructs an Event and increments the sequence counter.
func (s *GameState) NewEvent(typ string, payload any) Event {
	raw, _ := marshalJSON(payload)
	s.EventSeq++
	return Event{Type: typ, Payload: raw}
}
