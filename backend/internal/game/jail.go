package game

import "fmt"

const maxJailTurns = 3

// EnterJail sends the active player to jail.
func EnterJail(s *GameState) []Event {
	p := s.ActivePlayer()
	p.Position = PosJail
	p.InJail = true
	p.JailTurns = 0
	s.DoublesCount = 0

	return []Event{s.NewEvent("SENT_TO_JAIL", map[string]any{
		"playerId": p.BastionID,
	})}
}

// JailPayFine pays the $50 fine to exit jail.
func JailPayFine(s *GameState, rules *RuleSet) ([]Event, error) {
	p := s.ActivePlayer()
	if !p.InJail {
		return nil, fmt.Errorf("not in jail")
	}
	fine := 50
	if p.Balance < fine {
		return nil, fmt.Errorf("insufficient funds to pay fine")
	}
	p.Balance -= fine
	p.InJail = false
	p.JailTurns = 0

	return []Event{s.NewEvent("JAIL_FINE_PAID", map[string]any{
		"playerId": p.BastionID,
		"amount":   fine,
	})}, nil
}

// JailUseCard uses a get-out-of-jail-free card.
func JailUseCard(s *GameState) ([]Event, error) {
	p := s.ActivePlayer()
	if !p.InJail {
		return nil, fmt.Errorf("not in jail")
	}
	if p.JailCards <= 0 {
		return nil, fmt.Errorf("no get-out-of-jail card")
	}
	p.JailCards--
	p.InJail = false
	p.JailTurns = 0

	return []Event{s.NewEvent("JAIL_CARD_USED", map[string]any{
		"playerId": p.BastionID,
	})}, nil
}

// JailRoll attempts to roll doubles to exit jail. Advances jail turn counter.
// If 3 turns pass without doubles the fine is forced automatically.
func JailRoll(s *GameState, rules *RuleSet) (RollResult, []Event, error) {
	p := s.ActivePlayer()
	if !p.InJail {
		return RollResult{}, nil, fmt.Errorf("not in jail")
	}

	result, rollEvents, err := Roll(s)
	if err != nil {
		return RollResult{}, nil, err
	}

	var events []Event
	events = append(events, rollEvents...)

	if result.Doubles {
		p.InJail = false
		p.JailTurns = 0
		events = append(events, s.NewEvent("JAIL_EXIT_DOUBLES", map[string]any{
			"playerId": p.BastionID,
		}))
		events = append(events, MovePlayer(s, rules, result.Total)...)
		return result, events, nil
	}

	p.JailTurns++
	if p.JailTurns >= maxJailTurns {
		// Force fine on third failed roll
		fine := 50
		p.Balance -= fine
		p.InJail = false
		p.JailTurns = 0
		events = append(events, s.NewEvent("JAIL_FINE_FORCED", map[string]any{
			"playerId": p.BastionID,
			"amount":   fine,
		}))
		events = append(events, MovePlayer(s, rules, result.Total)...)
	}

	return result, events, nil
}
