package game

import (
	"encoding/json"
	"fmt"
)

// ApplyCard executes a drawn card's effect against the game state.
func ApplyCard(s *GameState, rules *RuleSet, card Card) ([]Event, error) {
	var events []Event

	events = append(events, s.NewEvent("CARD_DRAWN", map[string]any{
		"playerId":    s.ActivePlayer().BastionID,
		"deck":        card.Deck,
		"name":        card.Name,
		"description": card.Description,
		"effectType":  card.EffectType,
	}))

	var err error
	var cardEvents []Event

	switch card.EffectType {
	case EffectMoveTo:
		cardEvents, err = applyMoveTo(s, rules, card.EffectPayload)
	case EffectMoveRelative:
		cardEvents, err = applyMoveRelative(s, rules, card.EffectPayload)
	case EffectNearestRailroad:
		cardEvents, err = applyNearestRailroad(s, rules, card.EffectPayload)
	case EffectNearestUtility:
		cardEvents, err = applyNearestUtility(s, rules, card.EffectPayload)
	case EffectCollect:
		cardEvents, err = applyCollect(s, card.EffectPayload)
	case EffectPay:
		cardEvents, err = applyPay(s, rules, card.EffectPayload)
	case EffectCollectPerPlayer:
		cardEvents, err = applyCollectPerPlayer(s, card.EffectPayload)
	case EffectPayPerPlayer:
		cardEvents, err = applyPayPerPlayer(s, card.EffectPayload)
	case EffectPayPerBuilding:
		cardEvents, err = applyPayPerBuilding(s, rules, card.EffectPayload)
	case EffectCollectPerBuilding:
		cardEvents, err = applyCollectPerBuilding(s, rules, card.EffectPayload)
	case EffectGoToJail:
		cardEvents = EnterJail(s)
	case EffectGetOutOfJail:
		s.ActivePlayer().JailCards++
		cardEvents = []Event{s.NewEvent("JAIL_CARD_RECEIVED", map[string]any{
			"playerId": s.ActivePlayer().BastionID,
		})}
	case EffectBackToGo:
		s.ActivePlayer().Position = PosGO
		cardEvents = []Event{s.NewEvent("PLAYER_MOVED", map[string]any{
			"playerId": s.ActivePlayer().BastionID,
			"to":       PosGO,
		})}
	default:
		err = fmt.Errorf("unknown effect type: %s", card.EffectType)
	}

	if err != nil {
		return events, err
	}
	return append(events, cardEvents...), nil
}

// ---- effect handlers ----

func applyMoveTo(s *GameState, rules *RuleSet, payload json.RawMessage) ([]Event, error) {
	var p struct {
		Position   int  `json:"position"`
		CollectGO  bool `json:"collect_go"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	return MovePlayerTo(s, rules, p.Position, p.CollectGO), nil
}

func applyMoveRelative(s *GameState, rules *RuleSet, payload json.RawMessage) ([]Event, error) {
	var p struct {
		Steps int `json:"steps"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	if p.Steps >= 0 {
		return MovePlayer(s, rules, p.Steps), nil
	}
	// Backward move: no GO collection
	player := s.ActivePlayer()
	old := player.Position
	player.Position = (old + p.Steps + 40) % 40
	return []Event{s.NewEvent("PLAYER_MOVED", map[string]any{
		"playerId": player.BastionID,
		"from":     old,
		"to":       player.Position,
		"steps":    p.Steps,
	})}, nil
}

func applyNearestRailroad(s *GameState, rules *RuleSet, payload json.RawMessage) ([]Event, error) {
	var p struct {
		DoubleRent bool `json:"double_rent"`
	}
	_ = json.Unmarshal(payload, &p)

	railroads := tilesOfType(rules, TileRailroad)
	target, _ := NearestPosition(s.ActivePlayer().Position, railroads)
	events := MovePlayerTo(s, rules, target, true)

	// If owned and double_rent, pay double rent
	if p.DoubleRent {
		ts := s.TileAt(target)
		if ts.OwnerID != nil && !ts.Mortgaged {
			rent := CalculateRent(s, rules, target, 0) * 2
			payer := s.ActivePlayer()
			owner := s.OwnerOf(target)
			if owner != nil && owner.BastionID != payer.BastionID {
				payer.Balance -= rent
				owner.Balance += rent
				events = append(events, s.NewEvent("RENT_PAID", map[string]any{
					"from": payer.BastionID, "to": owner.BastionID,
					"tilePosition": target, "amount": rent, "doubled": true,
				}))
			}
		}
	}
	return events, nil
}

func applyNearestUtility(s *GameState, rules *RuleSet, payload json.RawMessage) ([]Event, error) {
	var p struct {
		DiceMult int `json:"dice_mult"`
	}
	_ = json.Unmarshal(payload, &p)

	utilities := tilesOfType(rules, TileUtility)
	target, _ := NearestPosition(s.ActivePlayer().Position, utilities)
	events := MovePlayerTo(s, rules, target, true)

	ts := s.TileAt(target)
	if ts.OwnerID != nil && !ts.Mortgaged && p.DiceMult > 0 {
		diceTotal := s.Dice[0] + s.Dice[1]
		rent := p.DiceMult * diceTotal
		payer := s.ActivePlayer()
		owner := s.OwnerOf(target)
		if owner != nil && owner.BastionID != payer.BastionID {
			payer.Balance -= rent
			owner.Balance += rent
			events = append(events, s.NewEvent("RENT_PAID", map[string]any{
				"from": payer.BastionID, "to": owner.BastionID,
				"tilePosition": target, "amount": rent,
			}))
		}
	}
	return events, nil
}

func applyCollect(s *GameState, payload json.RawMessage) ([]Event, error) {
	var p struct{ Amount int `json:"amount"` }
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	player := s.ActivePlayer()
	player.Balance += p.Amount
	return []Event{s.NewEvent("BANK_PAID_PLAYER", map[string]any{
		"playerId": player.BastionID, "amount": p.Amount,
	})}, nil
}

func applyPay(s *GameState, rules *RuleSet, payload json.RawMessage) ([]Event, error) {
	var p struct{ Amount int `json:"amount"` }
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	player := s.ActivePlayer()

	if rules.Config.FreeParkingJackpot {
		s.FreeParkingPot += p.Amount
	}
	player.Balance -= p.Amount
	return []Event{s.NewEvent("PLAYER_PAID_BANK", map[string]any{
		"playerId": player.BastionID, "amount": p.Amount,
	})}, nil
}

func applyCollectPerPlayer(s *GameState, payload json.RawMessage) ([]Event, error) {
	var p struct{ Amount int `json:"amount"` }
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	collector := s.ActivePlayer()
	total := 0
	for i := range s.Players {
		if s.Players[i].BastionID == collector.BastionID || s.Players[i].Bankrupt {
			continue
		}
		s.Players[i].Balance -= p.Amount
		total += p.Amount
	}
	collector.Balance += total
	return []Event{s.NewEvent("COLLECTED_FROM_PLAYERS", map[string]any{
		"playerId": collector.BastionID, "perPlayer": p.Amount, "total": total,
	})}, nil
}

func applyPayPerPlayer(s *GameState, payload json.RawMessage) ([]Event, error) {
	var p struct{ Amount int `json:"amount"` }
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	payer := s.ActivePlayer()
	for i := range s.Players {
		if s.Players[i].BastionID == payer.BastionID || s.Players[i].Bankrupt {
			continue
		}
		payer.Balance -= p.Amount
		s.Players[i].Balance += p.Amount
	}
	return []Event{s.NewEvent("PAID_TO_PLAYERS", map[string]any{
		"playerId": payer.BastionID, "perPlayer": p.Amount,
	})}, nil
}

func applyPayPerBuilding(s *GameState, rules *RuleSet, payload json.RawMessage) ([]Event, error) {
	var p struct {
		House int `json:"house"`
		Hotel int `json:"hotel"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	player := s.ActivePlayer()
	total := buildingRepairCost(s, rules, player, p.House, p.Hotel)

	if rules.Config.FreeParkingJackpot {
		s.FreeParkingPot += total
	}
	player.Balance -= total
	return []Event{s.NewEvent("BUILDING_REPAIRS_PAID", map[string]any{
		"playerId":   player.BastionID,
		"houseCost":  p.House,
		"hotelCost":  p.Hotel,
		"total":      total,
	})}, nil
}

func applyCollectPerBuilding(s *GameState, rules *RuleSet, payload json.RawMessage) ([]Event, error) {
	var p struct {
		House int `json:"house"`
		Hotel int `json:"hotel"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	player := s.ActivePlayer()
	total := buildingRepairCost(s, rules, player, p.House, p.Hotel)
	player.Balance += total
	return []Event{s.NewEvent("BUILDING_COLLECTED", map[string]any{
		"playerId": player.BastionID, "total": total,
	})}, nil
}

func buildingRepairCost(s *GameState, rules *RuleSet, player *Player, houseCost, hotelCost int) int {
	total := 0
	for pos := range rules.Props {
		ts := s.TileAt(pos)
		if ts.OwnerID == nil || *ts.OwnerID != player.ID {
			continue
		}
		if ts.Hotel {
			total += hotelCost
		} else {
			total += ts.Houses * houseCost
		}
	}
	return total
}

func tilesOfType(rules *RuleSet, tileType string) []int {
	var out []int
	for i, t := range rules.Tiles {
		if t.Type == tileType {
			out = append(out, i)
		}
	}
	return out
}
