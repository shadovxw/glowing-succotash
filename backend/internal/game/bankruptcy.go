package game

import "fmt"

// CanAfford returns true if the player has enough liquid + liquidatable assets
// to cover the debt. Used before forcing bankruptcy.
func CanAfford(s *GameState, rules *RuleSet, player *Player, amount int) bool {
	if player.Balance >= amount {
		return true
	}
	total := player.Balance + liquidationValue(s, rules, player)
	return total >= amount
}

// DeclareBankruptcy marks a player bankrupt and transfers all their assets
// to the creditor (another player) or back to the bank.
func DeclareBankruptcy(s *GameState, rules *RuleSet, bankruptID, creditorID string) ([]Event, error) {
	bankrupt := s.PlayerByBastionID(bankruptID)
	if bankrupt == nil {
		return nil, fmt.Errorf("player not found")
	}
	bankrupt.Bankrupt = true
	bankrupt.IsActive = false

	var events []Event

	if creditorID != "" {
		creditor := s.PlayerByBastionID(creditorID)
		if creditor == nil {
			return nil, fmt.Errorf("creditor not found")
		}
		// Transfer cash
		creditor.Balance += bankrupt.Balance
		bankrupt.Balance = 0

		// Transfer properties (unmortgage interest is waived in bankruptcy)
		for pos := range rules.Props {
			ts := s.TileAt(pos)
			if ts.OwnerID != nil && *ts.OwnerID == bankrupt.ID {
				ts.OwnerID = &creditor.ID
			}
		}

		// Transfer jail cards
		creditor.JailCards += bankrupt.JailCards
		bankrupt.JailCards = 0

		events = append(events, s.NewEvent("PLAYER_BANKRUPT", map[string]any{
			"playerId":   bankruptID,
			"creditorId": creditorID,
		}))
	} else {
		// No creditor (tax or card) — assets return to bank
		bankrupt.Balance = 0
		for pos := range rules.Props {
			ts := s.TileAt(pos)
			if ts.OwnerID != nil && *ts.OwnerID == bankrupt.ID {
				ts.OwnerID = nil
				ts.Houses = 0
				ts.Hotel = false
				ts.Mortgaged = false
			}
		}
		events = append(events, s.NewEvent("PLAYER_BANKRUPT", map[string]any{
			"playerId":   bankruptID,
			"creditorId": nil,
		}))
	}

	// Check win condition
	if s.ActivePlayerCount() == 1 {
		winner := findLastPlayer(s)
		s.Status = StatusEnded
		events = append(events, s.NewEvent("GAME_OVER", map[string]any{
			"winnerId": winner.BastionID,
		}))
	} else {
		s.AdvanceTurn()
	}

	return events, nil
}

// liquidationValue estimates cash from selling all buildings at half price
// and mortgaging all unmortgaged properties.
func liquidationValue(s *GameState, rules *RuleSet, player *Player) int {
	total := 0
	for pos, prop := range rules.Props {
		ts := s.TileAt(pos)
		if ts.OwnerID == nil || *ts.OwnerID != player.ID {
			continue
		}
		if ts.Hotel {
			total += prop.HouseCost / 2 // hotel sell value
		}
		total += ts.Houses * (prop.HouseCost / 2)
		if !ts.Mortgaged {
			total += prop.MortgageValue
		}
	}
	return total
}

func findLastPlayer(s *GameState) *Player {
	for i := range s.Players {
		if !s.Players[i].Bankrupt {
			return &s.Players[i]
		}
	}
	return nil
}
