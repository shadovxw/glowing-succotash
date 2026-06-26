package game

// ResolveLanding dispatches post-movement logic based on the tile the active
// player just landed on. Returns events plus a bool indicating whether the
// turn should auto-advance (false means waiting for player input: buy/decline).
func ResolveLanding(s *GameState, rules *RuleSet) (events []Event, awaitingInput bool) {
	p := s.ActivePlayer()
	tile := rules.Tiles[p.Position]

	switch tile.Type {
	case TileGO:
		// GO salary already collected during movement
	case TileStreet, TileRailroad, TileUtility:
		events, awaitingInput = resolveOwnable(s, rules, p.Position)
	case TileTax:
		events = resolveTax(s, rules, p.Position, tile)
	case TileChance:
		card := s.DrawChance(rules)
		var err error
		events, err = ApplyCard(s, rules, card)
		_ = err // card errors are non-fatal; broadcast the draw regardless
	case TileCommunityChest:
		card := s.DrawCommunity(rules)
		var err error
		events, err = ApplyCard(s, rules, card)
		_ = err
	case TileGoToJail:
		events = EnterJail(s)
	case TileFreeParking:
		if rules.Config.FreeParkingJackpot && s.FreeParkingPot > 0 {
			p.Balance += s.FreeParkingPot
			events = append(events, s.NewEvent("FREE_PARKING_COLLECTED", map[string]any{
				"playerId": p.BastionID,
				"amount":   s.FreeParkingPot,
			}))
			s.FreeParkingPot = 0
		}
	case TileJail:
		// Just visiting — nothing happens
	}

	return events, awaitingInput
}

func resolveOwnable(s *GameState, rules *RuleSet, pos int) ([]Event, bool) {
	ts := s.TileAt(pos)

	if ts.OwnerID == nil {
		// Unowned: prompt player to buy or decline
		s.Phase = PhasePostRoll
		prop := rules.Props[pos]
		return []Event{s.NewEvent("BUY_PROMPT", map[string]any{
			"tilePosition": pos,
			"price":        prop.Price,
		})}, true
	}

	if ts.Mortgaged {
		return nil, false
	}

	owner := s.OwnerOf(pos)
	if owner == nil || owner.BastionID == s.ActivePlayer().BastionID {
		return nil, false
	}

	rentEvents, _ := PayRent(s, rules, pos, s.Dice[0]+s.Dice[1])
	return rentEvents, false
}

func resolveTax(s *GameState, rules *RuleSet, pos int, tile Tile) []Event {
	p := s.ActivePlayer()
	var amount int

	// position 4 = income tax, position 38 = luxury tax (standard board)
	// but we read from config so it's flexible
	if pos == 4 {
		amount = rules.Config.IncomeTax
	} else {
		amount = rules.Config.LuxuryTax
	}

	if rules.Config.FreeParkingJackpot {
		s.FreeParkingPot += amount
	}
	p.Balance -= amount

	return []Event{s.NewEvent("TAX_PAID", map[string]any{
		"playerId": p.BastionID,
		"tile":     tile.Name,
		"amount":   amount,
	})}
}
