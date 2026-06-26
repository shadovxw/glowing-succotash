package game

// MovePlayer advances the active player by `steps` spaces, collecting GO
// salary if they pass or land on position 0.
func MovePlayer(s *GameState, rules *RuleSet, steps int) []Event {
	p := s.ActivePlayer()
	oldPos := p.Position
	newPos := (oldPos + steps) % 40

	passedGO := newPos < oldPos || steps >= 40

	p.Position = newPos

	var events []Event
	events = append(events, s.NewEvent("PLAYER_MOVED", map[string]any{
		"playerId": p.BastionID,
		"from":     oldPos,
		"to":       newPos,
		"steps":    steps,
	}))

	if passedGO && newPos != PosGO {
		events = append(events, collectGO(s, rules, p)...)
	} else if newPos == PosGO {
		amount := rules.Config.GoAmount
		if rules.Config.DoubleSalaryOnGo {
			amount *= 2
		}
		p.Balance += amount
		events = append(events, s.NewEvent("GO_COLLECTED", map[string]any{
			"playerId": p.BastionID,
			"amount":   amount,
		}))
	}

	return events
}

// MovePlayerTo teleports the active player to a specific position (used by
// cards). Collects GO salary only if the player actually passes GO.
func MovePlayerTo(s *GameState, rules *RuleSet, targetPos int, collectGOSalary bool) []Event {
	p := s.ActivePlayer()
	oldPos := p.Position

	steps := (targetPos - oldPos + 40) % 40
	if steps == 0 {
		steps = 40
	}

	p.Position = targetPos

	var events []Event
	events = append(events, s.NewEvent("PLAYER_MOVED", map[string]any{
		"playerId": p.BastionID,
		"from":     oldPos,
		"to":       targetPos,
		"steps":    steps,
	}))

	if collectGOSalary && targetPos != PosGO && steps < 40 && (targetPos < oldPos || steps > 0 && oldPos+steps >= 40) {
		events = append(events, collectGO(s, rules, p)...)
	}

	return events
}

// NearestPosition finds the nearest tile of a given set of positions ahead of
// the player (wrapping around the board). Returns the position and steps taken.
func NearestPosition(current int, candidates []int) (pos, steps int) {
	minSteps := 41
	nearest := candidates[0]
	for _, c := range candidates {
		s := (c - current + 40) % 40
		if s == 0 {
			s = 40
		}
		if s < minSteps {
			minSteps = s
			nearest = c
		}
	}
	return nearest, minSteps
}

func collectGO(s *GameState, rules *RuleSet, p *Player) []Event {
	amount := rules.Config.GoAmount
	p.Balance += amount
	return []Event{s.NewEvent("GO_COLLECTED", map[string]any{
		"playerId": p.BastionID,
		"amount":   amount,
	})}
}
