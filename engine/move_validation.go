package engine

func isLegalMove(state *GameState, move Move) bool {
	// bounds
	if move.From.File < 0 || move.From.File > 7 || move.From.Rank < 0 || move.From.Rank > 7 {
		return false
	}
	if move.To.File < 0 || move.To.File > 7 || move.To.Rank < 0 || move.To.Rank > 7 {
		return false
	}

	piece := state.Board.Squares[move.From.File][move.From.Rank]
	if piece == nil {
		return false
	}
	if piece.Color != state.Turn {
		return false
	}

	dest := state.Board.Squares[move.To.File][move.To.Rank]
	if dest != nil && dest.Color == piece.Color {
		return false // cannot capture own piece
	}

	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}
	sign := func(x int) int {
		if x < 0 {
			return -1
		}
		if x > 0 {
			return 1
		}
		return 0
	}

	pathClearBetween := func(a, b Position) bool {
		df := b.File - a.File
		dr := b.Rank - a.Rank
		sF := sign(df)
		sR := sign(dr)
		f := a.File + sF
		r := a.Rank + sR
		for f != b.File || r != b.Rank {
			if state.Board.Squares[f][r] != nil {
				return false
			}
			f += sF
			r += sR
		}
		return true
	}

	df := move.To.File - move.From.File
	dr := move.To.Rank - move.From.Rank
	adf := abs(df)
	adr := abs(dr)

	switch piece.Kind {
	case Pawn:
		if piece.Color == White {
			// forward one to empty
			if df == 0 && dr == 1 && dest == nil {
				return true
			}
			// initial double move from rank 1 to rank 3
			if df == 0 && dr == 2 && move.From.Rank == 1 && dest == nil {
				// ensure intermediate square is clear
				mid := Position{File: move.From.File, Rank: move.From.Rank + 1}
				if state.Board.Squares[mid.File][mid.Rank] == nil {
					return true
				}
			}
			// captures (including en passant)
			if adf == 1 && dr == 1 {
				if dest != nil && dest.Color == Black {
					return true
				}
				if dest == nil && state.HasEnPassant && move.To == state.EnPassant {
					return true
				}
			}
		} else {
			if df == 0 && dr == -1 && dest == nil {
				return true
			}
			// initial double move for black from rank 6 to rank 4
			if df == 0 && dr == -2 && move.From.Rank == 6 && dest == nil {
				mid := Position{File: move.From.File, Rank: move.From.Rank - 1}
				if state.Board.Squares[mid.File][mid.Rank] == nil {
					return true
				}
			}
			if adf == 1 && dr == -1 {
				if dest != nil && dest.Color == White {
					return true
				}
				if dest == nil && state.HasEnPassant && move.To == state.EnPassant {
					return true
				}
			}
		}
		return false
	case Knight:
		if (adf == 2 && adr == 1) || (adf == 1 && adr == 2) {
			return true
		}
		return false
	case Bishop:
		if adf == adr && adf > 0 && pathClearBetween(move.From, move.To) {
			return true
		}
		return false
	case Rook:
		if (df == 0 || dr == 0) && (adf+adr > 0) && pathClearBetween(move.From, move.To) {
			return true
		}
		return false
	case Queen:
		if ((df == 0 || dr == 0) || (adf == adr)) && (adf+adr > 0) && pathClearBetween(move.From, move.To) {
			return true
		}
		return false
	case King:
		// normal one-square king move
		if adf <= 1 && adr <= 1 {
			return true
		}

		// castling: king moves two squares horizontally
		if adr == 0 && adf == 2 {
			// determine side and rights
			if state.Turn == White {
				// white king rank 0 expected
				if move.From.Rank != 0 {
					return false
				}
				switch df {
				case 2:
					// king-side
					if !state.WhiteCanCastleKingSide {
						return false
					}
					// rook must be at h1
					rookPos := Position{File: 7, Rank: 0}
					rook := state.Board.Squares[rookPos.File][rookPos.Rank]
					if rook == nil || rook.Kind != Rook || rook.Color != White {
						return false
					}
					// path between king and destination must be clear and not attacked
					mid := Position{File: 5, Rank: 0}
					destPos := move.To
					if !pathClearBetween(move.From, destPos) {
						return false
					}
					if squareAttacked(state, move.From, opposite(state.Turn)) || squareAttacked(state, mid, opposite(state.Turn)) || squareAttacked(state, destPos, opposite(state.Turn)) {
						return false
					}
					return true
				case -2:
					// queen-side
					if !state.WhiteCanCastleQueenSide {
						return false
					}
					rookPos := Position{File: 0, Rank: 0}
					rook := state.Board.Squares[rookPos.File][rookPos.Rank]
					if rook == nil || rook.Kind != Rook || rook.Color != White {
						return false
					}
					mid1 := Position{File: 3, Rank: 0}
					destPos := move.To
					if !pathClearBetween(move.From, destPos) {
						return false
					}
					if squareAttacked(state, move.From, opposite(state.Turn)) || squareAttacked(state, mid1, opposite(state.Turn)) || squareAttacked(state, destPos, opposite(state.Turn)) {
						return false
					}
					return true
				}
			} else {
				// black
				if move.From.Rank != 7 {
					return false
				}
				switch df {
				case 2:
					if !state.BlackCanCastleKingSide {
						return false
					}
					rookPos := Position{File: 7, Rank: 7}
					rook := state.Board.Squares[rookPos.File][rookPos.Rank]
					if rook == nil || rook.Kind != Rook || rook.Color != Black {
						return false
					}
					mid := Position{File: 5, Rank: 7}
					destPos := move.To
					if !pathClearBetween(move.From, destPos) {
						return false
					}
					if squareAttacked(state, move.From, opposite(state.Turn)) || squareAttacked(state, mid, opposite(state.Turn)) || squareAttacked(state, destPos, opposite(state.Turn)) {
						return false
					}
					return true
				case -2:
					if !state.BlackCanCastleQueenSide {
						return false
					}
					rookPos := Position{File: 0, Rank: 7}
					rook := state.Board.Squares[rookPos.File][rookPos.Rank]
					if rook == nil || rook.Kind != Rook || rook.Color != Black {
						return false
					}
					mid1 := Position{File: 3, Rank: 7}
					destPos := move.To
					if !pathClearBetween(move.From, destPos) {
						return false
					}
					if squareAttacked(state, move.From, opposite(state.Turn)) || squareAttacked(state, mid1, opposite(state.Turn)) || squareAttacked(state, destPos, opposite(state.Turn)) {
						return false
					}
					return true
				}
			}
		}
		return false
	}

	return false
}
