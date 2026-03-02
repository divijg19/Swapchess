package engine

// IsInCheck reports whether the given side's king is currently attacked.
func IsInCheck(state *GameState, color Color) bool {
	kingPos, ok := findKingPosition(state, color)
	if !ok {
		// Invalid state (missing king) is treated as "in check".
		return true
	}
	return squareAttacked(state, kingPos, opposite(color))
}

func hasAnyLegalMove(state *GameState, color Color) bool {
	tmp := state.Clone()
	tmp.Turn = color

	for f := 0; f < 8; f++ {
		for r := 0; r < 8; r++ {
			p := tmp.Board.Squares[f][r]
			if p == nil || p.Color != color {
				continue
			}
			from := Position{File: f, Rank: r}
			for tf := 0; tf < 8; tf++ {
				for tr := 0; tr < 8; tr++ {
					to := Position{File: tf, Rank: tr}
					if from == to {
						continue
					}
					if isLegalMove(tmp, Move{From: from, To: to}) {
						return true
					}
				}
			}
		}
	}
	return false
}

// IsCheckmate reports whether the side to move is checkmated.
func IsCheckmate(state *GameState) bool {
	color := state.Turn
	return IsInCheck(state, color) && !hasAnyLegalMove(state, color)
}

// IsStalemate reports whether the side to move has no legal moves while not in check.
func IsStalemate(state *GameState) bool {
	color := state.Turn
	return !IsInCheck(state, color) && !hasAnyLegalMove(state, color)
}
