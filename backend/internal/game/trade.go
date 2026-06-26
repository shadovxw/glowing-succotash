package game

import (
	"fmt"

	"github.com/google/uuid"
)

// ProposeTrade creates a pending trade between two players.
func ProposeTrade(s *GameState, rules *RuleSet, proposerID, receiverID string, offer, request TradeOffer) ([]Event, error) {
	proposer := s.PlayerByBastionID(proposerID)
	receiver := s.PlayerByBastionID(receiverID)
	if proposer == nil || receiver == nil {
		return nil, fmt.Errorf("invalid player")
	}
	if err := validateOffer(s, offer, proposer); err != nil {
		return nil, fmt.Errorf("invalid offer: %w", err)
	}
	if err := validateOffer(s, request, receiver); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	// Block trading a property that has buildings on its color group
	if err := noBuildingsOnGroup(s, rules, offer.Properties); err != nil {
		return nil, err
	}
	if err := noBuildingsOnGroup(s, rules, request.Properties); err != nil {
		return nil, err
	}

	trade := Trade{
		ID:         uuid.New(),
		ProposerID: proposerID,
		ReceiverID: receiverID,
		Offer:      offer,
		Request:    request,
		Status:     "pending",
	}
	s.PendingTrades = append(s.PendingTrades, trade)
	s.Phase = PhaseTrade

	return []Event{s.NewEvent("TRADE_PROPOSED", map[string]any{
		"trade": trade,
	})}, nil
}

// AcceptTrade executes a pending trade.
func AcceptTrade(s *GameState, tradeID uuid.UUID, acceptorID string) ([]Event, error) {
	trade, idx := findTrade(s, tradeID)
	if trade == nil {
		return nil, fmt.Errorf("trade not found")
	}
	if trade.ReceiverID != acceptorID {
		return nil, fmt.Errorf("not your trade to accept")
	}

	proposer := s.PlayerByBastionID(trade.ProposerID)
	receiver := s.PlayerByBastionID(trade.ReceiverID)

	// Transfer cash
	proposer.Balance -= trade.Offer.Cash
	receiver.Balance += trade.Offer.Cash
	receiver.Balance -= trade.Request.Cash
	proposer.Balance += trade.Request.Cash

	// Transfer properties (offer: proposer → receiver)
	for _, pos := range trade.Offer.Properties {
		ts := s.TileAt(pos)
		ts.OwnerID = &receiver.ID
	}
	// Transfer properties (request: receiver → proposer)
	for _, pos := range trade.Request.Properties {
		ts := s.TileAt(pos)
		ts.OwnerID = &proposer.ID
	}

	// Transfer jail cards
	proposer.JailCards -= trade.Offer.JailCards
	receiver.JailCards += trade.Offer.JailCards
	receiver.JailCards -= trade.Request.JailCards
	proposer.JailCards += trade.Request.JailCards

	s.PendingTrades = append(s.PendingTrades[:idx], s.PendingTrades[idx+1:]...)
	if len(s.PendingTrades) == 0 {
		s.Phase = PhaseRoll
	}

	return []Event{s.NewEvent("TRADE_ACCEPTED", map[string]any{
		"tradeId":    tradeID,
		"proposerId": trade.ProposerID,
		"receiverId": trade.ReceiverID,
	})}, nil
}

// RejectTrade cancels a pending trade.
func RejectTrade(s *GameState, tradeID uuid.UUID, rejectorID string) ([]Event, error) {
	trade, idx := findTrade(s, tradeID)
	if trade == nil {
		return nil, fmt.Errorf("trade not found")
	}
	if trade.ReceiverID != rejectorID && trade.ProposerID != rejectorID {
		return nil, fmt.Errorf("not your trade")
	}

	s.PendingTrades = append(s.PendingTrades[:idx], s.PendingTrades[idx+1:]...)
	if len(s.PendingTrades) == 0 {
		s.Phase = PhaseRoll
	}

	return []Event{s.NewEvent("TRADE_REJECTED", map[string]any{
		"tradeId": tradeID,
	})}, nil
}

func validateOffer(s *GameState, offer TradeOffer, player *Player) error {
	if offer.Cash < 0 || offer.Cash > player.Balance {
		return fmt.Errorf("invalid cash amount")
	}
	if offer.JailCards < 0 || offer.JailCards > player.JailCards {
		return fmt.Errorf("insufficient jail cards")
	}
	for _, pos := range offer.Properties {
		ts := s.TileAt(pos)
		if ts.OwnerID == nil || *ts.OwnerID != player.ID {
			return fmt.Errorf("player does not own property at %d", pos)
		}
	}
	return nil
}

func noBuildingsOnGroup(s *GameState, rules *RuleSet, positions []int) error {
	for _, pos := range positions {
		group := rules.Tiles[pos].ColorGroup
		if group == "" {
			continue
		}
		for p2, prop := range rules.Props {
			if rules.Tiles[p2].ColorGroup != group {
				continue
			}
			ts := s.TileAt(prop.TilePosition)
			if ts.Houses > 0 || ts.Hotel {
				return fmt.Errorf("sell all buildings in the %s group before trading", group)
			}
		}
	}
	return nil
}

func findTrade(s *GameState, id uuid.UUID) (*Trade, int) {
	for i, t := range s.PendingTrades {
		if t.ID == id {
			return &s.PendingTrades[i], i
		}
	}
	return nil, -1
}
