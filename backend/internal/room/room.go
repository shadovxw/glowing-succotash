package room

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/shadovxw/monopoly/internal/game"
	"github.com/shadovxw/monopoly/internal/proto"
	"github.com/shadovxw/monopoly/internal/store"
	"gorm.io/gorm"
)

// Action is sent from a seat's read goroutine into the room actor.
type Action struct {
	SenderID string
	Msg      proto.Envelope
	Reply    chan<- error // optional; nil for fire-and-forget
}

// Room owns one game and all connected seats. Only the actor goroutine
// mutates State and Seats — no locks needed.
type Room struct {
	ID      string
	Seats   []*Seat
	State   *game.GameState
	Rules   *game.RuleSet
	actions chan Action
	db      *gorm.DB
}

const maxSeats = 6

func NewRoom(id string, rules *game.RuleSet, db *gorm.DB) *Room {
	r := &Room{
		ID:      id,
		Rules:   rules,
		Seats:   make([]*Seat, 0, maxSeats),
		actions: make(chan Action, 128),
		db:      db,
	}
	go r.actor()
	return r
}

// Join adds a player to the room or reclaims their seat after reconnect.
func (r *Room) Join(bastionID, displayName, avatar string) (*Seat, error) {
	ch := make(chan error, 1)
	var seat *Seat
	r.actions <- Action{
		SenderID: bastionID,
		Msg:      proto.Envelope{Type: "_join"},
		Reply:    ch,
	}
	// We need the seat back — use a separate result channel attached via context.
	// Simplified: join is handled synchronously in the actor but we expose it here.
	// For the full implementation the actor writes *Seat to a result channel.
	_ = ch
	_ = seat
	return nil, fmt.Errorf("use JoinSync")
}

// JoinSync is the synchronous version called from the WS handler before the
// read loop starts.
func (r *Room) JoinSync(bastionID, displayName, avatar string) (*Seat, error) {
	// Find existing seat (reconnect)
	for _, s := range r.Seats {
		if s.BastionID == bastionID {
			return s, nil
		}
	}
	if len(r.Seats) >= maxSeats {
		return nil, fmt.Errorf("room full")
	}
	seat := NewSeat(len(r.Seats), bastionID, displayName, avatar)
	r.Seats = append(r.Seats, seat)
	return seat, nil
}

// Dispatch sends an action into the room actor from a WS read goroutine.
func (r *Room) Dispatch(senderID string, msg proto.Envelope) {
	r.actions <- Action{SenderID: senderID, Msg: msg}
}

// Broadcast sends a message to all connected seats.
func (r *Room) Broadcast(data []byte) {
	for _, s := range r.Seats {
		if s.Connected() {
			select {
			case s.Send <- data:
			default:
			}
		}
	}
}

// Send delivers a message to one specific seat.
func (r *Room) Send(bastionID string, data []byte) {
	for _, s := range r.Seats {
		if s.BastionID == bastionID && s.Connected() {
			select {
			case s.Send <- data:
			default:
			}
		}
	}
}

// actor is the single goroutine that owns game state. It serialises all
// mutations so there are no races on GameState.
func (r *Room) actor() {
	for action := range r.actions {
		if err := r.handle(action); err != nil {
			errMsg, _ := proto.Encode(proto.MsgActionError, map[string]string{"error": err.Error()})
			r.Send(action.SenderID, errMsg)
		}
	}
}

func (r *Room) handle(action Action) error {
	var (
		events []Event
		err    error
	)

	senderID := action.SenderID
	msg := action.Msg

	switch msg.Type {
	case proto.MsgRollDice:
		events, err = r.handleRoll(senderID)
	case proto.MsgBuyProperty:
		events, err = r.handleBuyProperty(senderID)
	case proto.MsgDeclineProperty:
		events, err = r.handleDeclineProperty(senderID)
	case proto.MsgBuildHouse:
		events, err = r.handleBuild(senderID, msg.Payload, "house")
	case proto.MsgSellHouse:
		events, err = r.handleSell(senderID, msg.Payload, "house")
	case proto.MsgBuildHotel:
		events, err = r.handleBuild(senderID, msg.Payload, "hotel")
	case proto.MsgSellHotel:
		events, err = r.handleSell(senderID, msg.Payload, "hotel")
	case proto.MsgMortgage:
		events, err = r.handleMortgage(senderID, msg.Payload)
	case proto.MsgUnmortgage:
		events, err = r.handleUnmortgage(senderID, msg.Payload)
	case proto.MsgJailPay:
		events, err = r.handleJailPay(senderID)
	case proto.MsgJailUseCard:
		events, err = r.handleJailUseCard(senderID)
	case proto.MsgJailRoll:
		events, err = r.handleJailRoll(senderID)
	case proto.MsgEndTurn:
		events, err = r.handleEndTurn(senderID)
	case proto.MsgTradePropose:
		events, err = r.handleTradePropose(senderID, msg.Payload)
	case proto.MsgTradeAccept:
		events, err = r.handleTradeAccept(senderID, msg.Payload)
	case proto.MsgTradeReject:
		events, err = r.handleTradeReject(senderID, msg.Payload)
	case proto.MsgAuctionBid:
		events, err = r.handleAuctionBid(senderID, msg.Payload)
	case proto.MsgAuctionResolve:
		events, err = r.handleAuctionResolve(senderID)
	case proto.MsgDeclareBankruptcy:
		events, err = r.handleBankruptcy(senderID, msg.Payload)
	case proto.MsgChat:
		return r.handleChat(senderID, msg.Payload)
	case proto.MsgWebRTCOffer, proto.MsgWebRTCAnswer, proto.MsgWebRTCIce:
		return r.handleSignaling(senderID, msg)
	}

	if err != nil {
		return err
	}

	r.broadcastEvents(events)
	if r.State != nil {
		go store.SaveState(r.db, r.State) //nolint:errcheck
	}
	return nil
}

// broadcastEvents sends each game event as a WS message to all players.
func (r *Room) broadcastEvents(events []Event) {
	for _, ev := range events {
		data, err := proto.Encode(ev.Type, ev.Payload)
		if err != nil {
			log.Println("encode event:", err)
			continue
		}
		r.Broadcast(data)
	}
}

// ---- action handlers ----

func (r *Room) assertActive(senderID string) error {
	if r.State == nil {
		return fmt.Errorf("game not started")
	}
	if r.State.ActivePlayer().BastionID != senderID {
		return fmt.Errorf("not your turn")
	}
	return nil
}

func (r *Room) handleRoll(senderID string) ([]Event, error) {
	if err := r.assertActive(senderID); err != nil {
		return nil, err
	}
	p := r.State.ActivePlayer()
	if p.InJail {
		return nil, fmt.Errorf("use JAIL_ROLL when in jail")
	}
	if r.State.DoublesCount == 3 {
		return game.EnterJail(r.State), nil
	}

	result, rollEvents, err := game.Roll(r.State)
	if err != nil {
		return nil, err
	}

	events := toEvents(rollEvents)

	// Three doubles → jail (checked after recording this roll)
	if r.State.DoublesCount >= 3 {
		jailEvents := game.EnterJail(r.State)
		return append(events, toEvents(jailEvents)...), nil
	}

	moveEvents := game.MovePlayer(r.State, r.Rules, result.Total)
	events = append(events, toEvents(moveEvents)...)

	landEvents, awaitingInput := game.ResolveLanding(r.State, r.Rules)
	events = append(events, toEvents(landEvents)...)

	if !awaitingInput && !result.Doubles {
		r.State.AdvanceTurn()
	}

	return events, nil
}

func (r *Room) handleBuyProperty(senderID string) ([]Event, error) {
	if err := r.assertActive(senderID); err != nil {
		return nil, err
	}
	pos := r.State.ActivePlayer().Position
	buyEvents, err := game.BuyProperty(r.State, r.Rules, pos)
	if err != nil {
		return nil, err
	}
	r.State.Phase = game.PhaseRoll
	return toEvents(buyEvents), nil
}

func (r *Room) handleDeclineProperty(senderID string) ([]Event, error) {
	if err := r.assertActive(senderID); err != nil {
		return nil, err
	}
	pos := r.State.ActivePlayer().Position
	if !r.Rules.Config.AuctionOnDecline {
		r.State.Phase = game.PhaseRoll
		return nil, nil
	}
	auctionEvents, err := game.StartAuction(r.State, pos)
	return toEvents(auctionEvents), err
}

func (r *Room) handleBuild(senderID string, payload json.RawMessage, kind string) ([]Event, error) {
	if r.State == nil {
		return nil, fmt.Errorf("game not started")
	}
	var p proto.TilePositionPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	var evs []game.Event
	var err error
	if kind == "house" {
		evs, err = game.BuildHouse(r.State, r.Rules, p.TilePosition)
	} else {
		evs, err = game.BuildHotel(r.State, r.Rules, p.TilePosition)
	}
	return toEvents(evs), err
}

func (r *Room) handleSell(senderID string, payload json.RawMessage, kind string) ([]Event, error) {
	if r.State == nil {
		return nil, fmt.Errorf("game not started")
	}
	var p proto.TilePositionPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	var evs []game.Event
	var err error
	if kind == "house" {
		evs, err = game.SellHouse(r.State, r.Rules, p.TilePosition)
	} else {
		evs, err = game.SellHotel(r.State, r.Rules, p.TilePosition)
	}
	return toEvents(evs), err
}

func (r *Room) handleMortgage(senderID string, payload json.RawMessage) ([]Event, error) {
	if r.State == nil {
		return nil, fmt.Errorf("game not started")
	}
	var p proto.TilePositionPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	evs, err := game.MortgageProperty(r.State, r.Rules, p.TilePosition)
	return toEvents(evs), err
}

func (r *Room) handleUnmortgage(senderID string, payload json.RawMessage) ([]Event, error) {
	if r.State == nil {
		return nil, fmt.Errorf("game not started")
	}
	var p proto.TilePositionPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	evs, err := game.UnmortgageProperty(r.State, r.Rules, p.TilePosition)
	return toEvents(evs), err
}

func (r *Room) handleJailPay(senderID string) ([]Event, error) {
	if err := r.assertActive(senderID); err != nil {
		return nil, err
	}
	evs, err := game.JailPayFine(r.State, r.Rules)
	return toEvents(evs), err
}

func (r *Room) handleJailUseCard(senderID string) ([]Event, error) {
	if err := r.assertActive(senderID); err != nil {
		return nil, err
	}
	evs, err := game.JailUseCard(r.State)
	return toEvents(evs), err
}

func (r *Room) handleJailRoll(senderID string) ([]Event, error) {
	if err := r.assertActive(senderID); err != nil {
		return nil, err
	}
	_, evs, err := game.JailRoll(r.State, r.Rules)
	return toEvents(evs), err
}

func (r *Room) handleEndTurn(senderID string) ([]Event, error) {
	if err := r.assertActive(senderID); err != nil {
		return nil, err
	}
	r.State.AdvanceTurn()
	next := r.State.ActivePlayer()
	data, _ := proto.Encode(proto.MsgYourTurn, map[string]string{"playerId": next.BastionID})
	r.Send(next.BastionID, data)
	return nil, nil
}

func (r *Room) handleTradePropose(senderID string, payload json.RawMessage) ([]Event, error) {
	if r.State == nil {
		return nil, fmt.Errorf("game not started")
	}
	var p proto.TradePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	offer := game.TradeOffer{Cash: p.Offer.Cash, Properties: p.Offer.Properties, JailCards: p.Offer.JailCards}
	req := game.TradeOffer{Cash: p.Request.Cash, Properties: p.Request.Properties, JailCards: p.Request.JailCards}
	evs, err := game.ProposeTrade(r.State, r.Rules, senderID, p.ReceiverID, offer, req)
	return toEvents(evs), err
}

func (r *Room) handleTradeAccept(senderID string, payload json.RawMessage) ([]Event, error) {
	var p proto.TradePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	id, _ := uuid.Parse(p.TradeID)
	evs, err := game.AcceptTrade(r.State, id, senderID)
	return toEvents(evs), err
}

func (r *Room) handleTradeReject(senderID string, payload json.RawMessage) ([]Event, error) {
	var p proto.TradePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	id, _ := uuid.Parse(p.TradeID)
	evs, err := game.RejectTrade(r.State, id, senderID)
	return toEvents(evs), err
}

func (r *Room) handleAuctionBid(senderID string, payload json.RawMessage) ([]Event, error) {
	var p proto.AuctionBidPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	evs, err := game.PlaceBid(r.State, senderID, p.Amount)
	return toEvents(evs), err
}

func (r *Room) handleAuctionResolve(senderID string) ([]Event, error) {
	if err := r.assertActive(senderID); err != nil {
		return nil, err
	}
	evs, err := game.ResolveAuction(r.State)
	return toEvents(evs), err
}

func (r *Room) handleBankruptcy(senderID string, payload json.RawMessage) ([]Event, error) {
	var p proto.BankruptcyPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	evs, err := game.DeclareBankruptcy(r.State, r.Rules, senderID, p.CreditorID)
	if err != nil {
		return nil, err
	}
	if r.State.Status == game.StatusEnded {
		winner := findWinner(r.State)
		if winner != "" {
			_ = store.EndGame(r.db, r.State.ID, winner)
			_ = store.IncrementGamesWon(r.db, winner)
		}
		for _, s := range r.Seats {
			_ = store.IncrementGamesPlayed(r.db, s.BastionID)
		}
	}
	return toEvents(evs), nil
}

func (r *Room) handleChat(senderID string, payload json.RawMessage) error {
	var p proto.ChatPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	seat := r.seatByBastionID(senderID)
	if seat == nil {
		return fmt.Errorf("unknown sender")
	}

	var gameID *uuid.UUID
	if r.State != nil {
		id := r.State.ID
		gameID = &id
	}

	_ = store.InsertChat(r.db, r.ID, gameID, &senderID, seat.DisplayName, p.Message, false)

	data, _ := proto.Encode(proto.MsgChatMessage, map[string]any{
		"senderId":    senderID,
		"displayName": seat.DisplayName,
		"avatar":      seat.Avatar,
		"message":     p.Message,
	})
	r.Broadcast(data)
	return nil
}

func (r *Room) handleSignaling(senderID string, msg proto.Envelope) error {
	var p proto.WebRTCPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return err
	}
	// Forward the signal to the target peer only
	forward, _ := proto.Encode(msg.Type, map[string]any{
		"fromId": senderID,
		"data":   p.Data,
	})
	r.Send(p.TargetID, forward)
	return nil
}

func (r *Room) seatByBastionID(id string) *Seat {
	for _, s := range r.Seats {
		if s.BastionID == id {
			return s
		}
	}
	return nil
}

func findWinner(s *game.GameState) string {
	for _, p := range s.Players {
		if !p.Bankrupt {
			return p.BastionID
		}
	}
	return ""
}

// Event aliases game.Event for internal use
type Event = game.Event

func toEvents(evs []game.Event) []Event { return evs }
