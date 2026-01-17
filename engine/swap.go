package engine

import "math/rand"

func applySwap(state *GameState, movedPos Position) {
	movedPiece := state.Board.Squares[movedPos.File][movedPos.Rank]
	if movedPiece == nil {
		return
	}

	var candidates []Position

	for f := 0; f < 8; f++ {
		for r := 0; r < 8; r++ {
			if f == movedPos.File && r == movedPos.Rank {
				continue
			}
			p := state.Board.Squares[f][r]
			if p != nil && p.Color == movedPiece.Color {
				candidates = append(candidates, Position{f, r})
			}
		}
	}

	if len(candidates) == 0 {
		return
	}

	rng := rand.New(rand.NewSource(state.RandSeed))
	idx := rng.Intn(len(candidates))
	target := candidates[idx]

	// swap
	state.Board.Squares[target.File][target.Rank], state.Board.Squares[movedPos.File][movedPos.Rank] =
		state.Board.Squares[movedPos.File][movedPos.Rank], state.Board.Squares[target.File][target.Rank]

	state.RandSeed++
}
