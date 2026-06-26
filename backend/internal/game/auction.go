package game

import "fmt"

// StartAuction begins an auction for a property (called when a player declines to buy).
func StartAuction(s *GameState, tilePos int) ([]Event, error) {
	ts := s.TileAt(tilePos)
	if ts.OwnerID != nil {
		return nil, fmt.Errorf("property already owned")
	}

	s.ActiveAuction = &Auction{
		TilePosition: tilePos,
		Bids:         make(map[string]int),
	}
	s.Phase = PhaseAuction

	return []Event{s.NewEvent("AUCTION_STARTED", map[string]any{
		"tilePosition": tilePos,
	})}, nil
}

// PlaceBid records a bid from a player.
func PlaceBid(s *GameState, bidderID string, amount int) ([]Event, error) {
	if s.ActiveAuction == nil {
		return nil, fmt.Errorf("no active auction")
	}
	if amount <= s.ActiveAuction.HighestBid {
		return nil, fmt.Errorf("bid must exceed current highest bid of %d", s.ActiveAuction.HighestBid)
	}
	bidder := s.PlayerByBastionID(bidderID)
	if bidder == nil || bidder.Bankrupt {
		return nil, fmt.Errorf("invalid bidder")
	}
	if bidder.Balance < amount {
		return nil, fmt.Errorf("insufficient funds")
	}

	s.ActiveAuction.Bids[bidderID] = amount
	s.ActiveAuction.HighestBid = amount
	s.ActiveAuction.HighestBidder = bidderID

	return []Event{s.NewEvent("AUCTION_BID", map[string]any{
		"playerId": bidderID,
		"amount":   amount,
	})}, nil
}

// ResolveAuction closes the auction and transfers the property to the winner.
func ResolveAuction(s *GameState) ([]Event, error) {
	a := s.ActiveAuction
	if a == nil {
		return nil, fmt.Errorf("no active auction")
	}

	var events []Event

	if a.HighestBidder == "" {
		// No bids — property stays with bank
		events = append(events, s.NewEvent("AUCTION_NO_WINNER", map[string]any{
			"tilePosition": a.TilePosition,
		}))
	} else {
		winner := s.PlayerByBastionID(a.HighestBidder)
		winner.Balance -= a.HighestBid
		ts := s.TileAt(a.TilePosition)
		ts.OwnerID = &winner.ID

		events = append(events, s.NewEvent("AUCTION_WON", map[string]any{
			"tilePosition": a.TilePosition,
			"playerId":     a.HighestBidder,
			"amount":       a.HighestBid,
		}))
	}

	s.ActiveAuction = nil
	s.Phase = PhaseRoll
	return events, nil
}
