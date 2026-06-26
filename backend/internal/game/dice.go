package game

import (
	"fmt"
	"math/rand"
)

type RollResult struct {
	D1      int  `json:"d1"`
	D2      int  `json:"d2"`
	Total   int  `json:"total"`
	Doubles bool `json:"doubles"`
}

// Roll rolls two dice and updates state. Returns an error if it is not the
// roll phase or the player already rolled doubles twice (handled externally
// — caller checks three-doubles-to-jail before calling Roll again).
func Roll(s *GameState) (RollResult, []Event, error) {
	if s.Phase != PhaseRoll {
		return RollResult{}, nil, fmt.Errorf("not in roll phase")
	}

	d1 := rand.Intn(6) + 1
	d2 := rand.Intn(6) + 1
	result := RollResult{
		D1:      d1,
		D2:      d2,
		Total:   d1 + d2,
		Doubles: d1 == d2,
	}

	s.Dice = [2]int{d1, d2}

	if result.Doubles {
		s.DoublesCount++
	}

	ev := s.NewEvent("DICE_ROLLED", result)
	return result, []Event{ev}, nil
}
