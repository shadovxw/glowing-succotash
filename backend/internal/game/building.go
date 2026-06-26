package game

import "fmt"

// BuildHouse places one house on a street tile, enforcing all standard rules.
func BuildHouse(s *GameState, rules *RuleSet, tilePos int) ([]Event, error) {
	p := s.ActivePlayer()
	ts := s.TileAt(tilePos)

	if err := assertOwns(ts, p); err != nil {
		return nil, err
	}
	if ts.Mortgaged {
		return nil, fmt.Errorf("property is mortgaged")
	}
	if ts.Hotel {
		return nil, fmt.Errorf("already has a hotel")
	}
	if ts.Houses >= 4 {
		return nil, fmt.Errorf("must build hotel next")
	}
	if !ownsFullGroup(s, rules, tilePos, p) {
		return nil, fmt.Errorf("must own full color group to build")
	}
	if !canAddHouse(s, rules, tilePos, ts.Houses) {
		return nil, fmt.Errorf("must build evenly across color group")
	}

	prop := rules.Props[tilePos]
	if prop.HouseCost == 0 {
		return nil, fmt.Errorf("cannot build on this tile type")
	}
	if p.Balance < prop.HouseCost {
		return nil, fmt.Errorf("insufficient funds")
	}
	if s.HousesLeft <= 0 {
		return nil, fmt.Errorf("no houses left in the bank")
	}

	p.Balance -= prop.HouseCost
	ts.Houses++
	s.HousesLeft--

	return []Event{s.NewEvent("HOUSE_BUILT", map[string]any{
		"playerId":     p.BastionID,
		"tilePosition": tilePos,
		"houses":       ts.Houses,
		"cost":         prop.HouseCost,
	})}, nil
}

// SellHouse removes one house from a street tile, returning half the house cost.
func SellHouse(s *GameState, rules *RuleSet, tilePos int) ([]Event, error) {
	p := s.ActivePlayer()
	ts := s.TileAt(tilePos)

	if err := assertOwns(ts, p); err != nil {
		return nil, err
	}
	if ts.Houses <= 0 {
		return nil, fmt.Errorf("no houses to sell")
	}
	if !canRemoveHouse(s, rules, tilePos, ts.Houses) {
		return nil, fmt.Errorf("must sell evenly across color group")
	}

	prop := rules.Props[tilePos]
	refund := prop.HouseCost / 2
	p.Balance += refund
	ts.Houses--
	s.HousesLeft++

	return []Event{s.NewEvent("HOUSE_SOLD", map[string]any{
		"playerId":     p.BastionID,
		"tilePosition": tilePos,
		"houses":       ts.Houses,
		"refund":       refund,
	})}, nil
}

// BuildHotel upgrades 4 houses to a hotel, returning the houses to the bank.
func BuildHotel(s *GameState, rules *RuleSet, tilePos int) ([]Event, error) {
	p := s.ActivePlayer()
	ts := s.TileAt(tilePos)

	if err := assertOwns(ts, p); err != nil {
		return nil, err
	}
	if ts.Hotel {
		return nil, fmt.Errorf("already has a hotel")
	}
	if ts.Houses < 4 {
		return nil, fmt.Errorf("need 4 houses before building a hotel")
	}
	if !canAddHotel(s, rules, tilePos) {
		return nil, fmt.Errorf("must build evenly across color group")
	}

	prop := rules.Props[tilePos]
	if p.Balance < prop.HouseCost {
		return nil, fmt.Errorf("insufficient funds")
	}
	if s.HotelsLeft <= 0 {
		return nil, fmt.Errorf("no hotels left in the bank")
	}

	p.Balance -= prop.HouseCost
	s.HousesLeft += 4 // houses returned to bank
	ts.Houses = 0
	ts.Hotel = true
	s.HotelsLeft--

	return []Event{s.NewEvent("HOTEL_BUILT", map[string]any{
		"playerId":     p.BastionID,
		"tilePosition": tilePos,
		"cost":         prop.HouseCost,
	})}, nil
}

// SellHotel downgrades a hotel back to 4 houses (if available) or 0.
func SellHotel(s *GameState, rules *RuleSet, tilePos int) ([]Event, error) {
	p := s.ActivePlayer()
	ts := s.TileAt(tilePos)

	if err := assertOwns(ts, p); err != nil {
		return nil, err
	}
	if !ts.Hotel {
		return nil, fmt.Errorf("no hotel to sell")
	}

	prop := rules.Props[tilePos]
	refund := prop.HouseCost / 2

	// Try to give back 4 houses; if bank has fewer, give what's available.
	housesBack := 4
	if s.HousesLeft < 4 {
		housesBack = s.HousesLeft
	}

	ts.Hotel = false
	ts.Houses = housesBack
	s.HousesLeft -= housesBack
	s.HotelsLeft++
	p.Balance += refund

	return []Event{s.NewEvent("HOTEL_SOLD", map[string]any{
		"playerId":     p.BastionID,
		"tilePosition": tilePos,
		"housesBack":   housesBack,
		"refund":       refund,
	})}, nil
}

// canAddHouse checks the even-build rule: no second house until every lot has one.
func canAddHouse(s *GameState, rules *RuleSet, tilePos, currentHouses int) bool {
	group := rules.Tiles[tilePos].ColorGroup
	for pos, prop := range rules.Props {
		if rules.Tiles[pos].ColorGroup != group || pos == tilePos {
			continue
		}
		other := s.TileAt(prop.TilePosition)
		if other.Houses < currentHouses {
			return false
		}
	}
	return true
}

// canRemoveHouse checks the even-sell rule: no second removal until every lot has one removed.
func canRemoveHouse(s *GameState, rules *RuleSet, tilePos, currentHouses int) bool {
	group := rules.Tiles[tilePos].ColorGroup
	for pos, prop := range rules.Props {
		if rules.Tiles[pos].ColorGroup != group || pos == tilePos {
			continue
		}
		other := s.TileAt(prop.TilePosition)
		if other.Houses > currentHouses {
			return false
		}
	}
	return true
}

// canAddHotel checks that all other lots in the group already have 4 houses.
func canAddHotel(s *GameState, rules *RuleSet, tilePos int) bool {
	group := rules.Tiles[tilePos].ColorGroup
	for pos, prop := range rules.Props {
		if rules.Tiles[pos].ColorGroup != group || pos == tilePos {
			continue
		}
		other := s.TileAt(prop.TilePosition)
		if !other.Hotel && other.Houses < 4 {
			return false
		}
	}
	return true
}
