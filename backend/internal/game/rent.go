package game

// CalculateRent returns the rent owed when a player lands on a tile owned by
// someone else. Returns 0 if no rent is due (unowned, mortgaged, owner in jail
// with noRentInJail rule, or owned by the landing player).
func CalculateRent(s *GameState, rules *RuleSet, landedPos int, diceTotal int) int {
	ts := s.TileAt(landedPos)
	if ts.OwnerID == nil || ts.Mortgaged {
		return 0
	}

	owner := s.OwnerOf(landedPos)
	if owner == nil || owner.BastionID == s.ActivePlayer().BastionID {
		return 0
	}

	if rules.Config.NoRentInJail && owner.InJail {
		return 0
	}

	prop, ok := rules.Props[landedPos]
	if !ok {
		return 0
	}

	tile := rules.Tiles[landedPos]

	switch tile.Type {
	case TileStreet:
		return streetRent(s, rules, prop, ts, owner)
	case TileRailroad:
		return railroadRent(s, rules, prop, owner)
	case TileUtility:
		return utilityRent(s, rules, prop, owner, diceTotal)
	}
	return 0
}

func streetRent(s *GameState, rules *RuleSet, prop *Property, ts *TileState, owner *Player) int {
	if ts.Hotel {
		return prop.Rent[5]
	}
	if ts.Houses > 0 {
		return prop.Rent[ts.Houses] // index 1-4
	}
	// No buildings: check if owner has the full color group (rent doubles)
	if ownsFullGroup(s, rules, prop.TilePosition, owner) {
		return prop.Rent[0] * 2
	}
	return prop.Rent[0]
}

func railroadRent(s *GameState, rules *RuleSet, prop *Property, owner *Player) int {
	count := countOwnedInGroup(s, rules, owner, TileRailroad)
	if count < 1 || count > len(prop.Rent) {
		return 0
	}
	return prop.Rent[count-1]
}

func utilityRent(s *GameState, rules *RuleSet, prop *Property, owner *Player, diceTotal int) int {
	count := countOwnedInGroup(s, rules, owner, TileUtility)
	if count < 1 || count > len(prop.Rent) {
		return 0
	}
	return prop.Rent[count-1] * diceTotal
}

// ownsFullGroup returns true if the owner holds all tiles in the same color group.
func ownsFullGroup(s *GameState, rules *RuleSet, tilePos int, owner *Player) bool {
	group := rules.Tiles[tilePos].ColorGroup
	if group == "" {
		return false
	}
	for pos, prop := range rules.Props {
		if rules.Tiles[pos].ColorGroup != group {
			continue
		}
		ts := s.TileAt(prop.TilePosition)
		if ts.OwnerID == nil || *ts.OwnerID != owner.ID {
			return false
		}
	}
	return true
}

// countOwnedInGroup counts how many tiles of the given type are owned by the player.
func countOwnedInGroup(s *GameState, rules *RuleSet, owner *Player, tileType string) int {
	count := 0
	for pos := range rules.Props {
		if rules.Tiles[pos].Type != tileType {
			continue
		}
		ts := s.TileAt(pos)
		if ts.OwnerID != nil && *ts.OwnerID == owner.ID {
			count++
		}
	}
	return count
}
