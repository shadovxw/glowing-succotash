package game

import (
	"fmt"

	"github.com/google/uuid"
)

// BuyProperty transfers a property from the bank to the active player.
func BuyProperty(s *GameState, rules *RuleSet, tilePos int) ([]Event, error) {
	ts := s.TileAt(tilePos)
	if ts.OwnerID != nil {
		return nil, fmt.Errorf("property already owned")
	}
	prop, ok := rules.Props[tilePos]
	if !ok {
		return nil, fmt.Errorf("not a property")
	}
	p := s.ActivePlayer()
	if p.Balance < prop.Price {
		return nil, fmt.Errorf("insufficient funds")
	}

	p.Balance -= prop.Price
	ts.OwnerID = &p.ID

	return []Event{s.NewEvent("PROPERTY_BOUGHT", map[string]any{
		"playerId":     p.BastionID,
		"tilePosition": tilePos,
		"price":        prop.Price,
	})}, nil
}

// PayRent transfers rent from the active player to the owner.
func PayRent(s *GameState, rules *RuleSet, tilePos int, diceTotal int) ([]Event, error) {
	rent := CalculateRent(s, rules, tilePos, diceTotal)
	if rent == 0 {
		return nil, nil
	}

	payer := s.ActivePlayer()
	owner := s.OwnerOf(tilePos)

	payer.Balance -= rent
	owner.Balance += rent

	return []Event{s.NewEvent("RENT_PAID", map[string]any{
		"from":         payer.BastionID,
		"to":           owner.BastionID,
		"tilePosition": tilePos,
		"amount":       rent,
	})}, nil
}

// MortgageProperty mortgages a property the active player owns.
func MortgageProperty(s *GameState, rules *RuleSet, tilePos int) ([]Event, error) {
	ts := s.TileAt(tilePos)
	p := s.ActivePlayer()

	if err := assertOwns(ts, p); err != nil {
		return nil, err
	}
	if ts.Mortgaged {
		return nil, fmt.Errorf("already mortgaged")
	}
	if ts.Houses > 0 || ts.Hotel {
		return nil, fmt.Errorf("sell all buildings before mortgaging")
	}

	prop := rules.Props[tilePos]
	ts.Mortgaged = true
	p.Balance += prop.MortgageValue

	return []Event{s.NewEvent("PROPERTY_MORTGAGED", map[string]any{
		"playerId":     p.BastionID,
		"tilePosition": tilePos,
		"amount":       prop.MortgageValue,
	})}, nil
}

// UnmortgageProperty lifts a mortgage; costs mortgage value + 10% interest.
func UnmortgageProperty(s *GameState, rules *RuleSet, tilePos int) ([]Event, error) {
	ts := s.TileAt(tilePos)
	p := s.ActivePlayer()

	if err := assertOwns(ts, p); err != nil {
		return nil, err
	}
	if !ts.Mortgaged {
		return nil, fmt.Errorf("not mortgaged")
	}

	prop := rules.Props[tilePos]
	cost := prop.MortgageValue + (prop.MortgageValue / 10)
	if p.Balance < cost {
		return nil, fmt.Errorf("insufficient funds: need %d", cost)
	}

	ts.Mortgaged = false
	p.Balance -= cost

	return []Event{s.NewEvent("PROPERTY_UNMORTGAGED", map[string]any{
		"playerId":     p.BastionID,
		"tilePosition": tilePos,
		"cost":         cost,
	})}, nil
}

func assertOwns(ts *TileState, p *Player) error {
	if ts.OwnerID == nil {
		return fmt.Errorf("property is unowned")
	}
	ownerID := uuid.UUID(*ts.OwnerID)
	if ownerID != p.ID {
		return fmt.Errorf("you do not own this property")
	}
	return nil
}
