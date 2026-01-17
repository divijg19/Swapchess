package engine

import "errors"

type GameState struct {
	Board                   Board
	Turn                    Color
	SuppressNextSwap        bool
	RandSeed                int64
	HasEnPassant            bool
	EnPassant               Position
	WhiteCanCastleKingSide  bool
	WhiteCanCastleQueenSide bool
	BlackCanCastleKingSide  bool
	BlackCanCastleQueenSide bool
}

var (
	ErrIllegalMove = errors.New("illegal move")
)

func ApplyMove(state *GameState, move Move) error {
	// Step order matters. Do not reorder casually.
	if !isLegalMove(state, move) {
		return ErrIllegalMove
	}

	movedPiece := state.Board.Squares[move.From.File][move.From.Rank]

	// 1. Apply the move
	// detect en-passant capture before we overwrite squares
	dest := state.Board.Squares[move.To.File][move.To.Rank]
	isEnPassantCapture := false
	if movedPiece != nil && movedPiece.Kind == Pawn && dest == nil && state.HasEnPassant {
		if move.To == state.EnPassant {
			isEnPassantCapture = true
		}
	}

	// detect castling: king moving two squares horizontally
	isCastling := false
	var castlingSide string
	if movedPiece != nil && movedPiece.Kind == King {
		df := move.To.File - move.From.File
		switch df {
		case 2:
			isCastling = true
			castlingSide = "king"
		case -2:
			isCastling = true
			castlingSide = "queen"
		}
	}

	state.Board.Squares[move.To.File][move.To.Rank] = movedPiece
	state.Board.Squares[move.From.File][move.From.Rank] = nil

	// handle en-passant captured pawn removal
	if isEnPassantCapture {
		// captured pawn sits on same file as move.To, at the pawn's origin rank
		captured := Position{File: move.To.File, Rank: move.From.Rank}
		state.Board.Squares[captured.File][captured.Rank] = nil
	}

	// handle castling rook movement
	if isCastling && movedPiece != nil && movedPiece.Kind == King {
		if movedPiece.Color == White {
			if castlingSide == "king" {
				// move rook from h1 to f1
				rook := state.Board.Squares[7][0]
				state.Board.Squares[5][0] = rook
				state.Board.Squares[7][0] = nil
			} else {
				// queen-side: move rook from a1 to d1
				rook := state.Board.Squares[0][0]
				state.Board.Squares[3][0] = rook
				state.Board.Squares[0][0] = nil
			}
		} else {
			if castlingSide == "king" {
				// black king-side: h8 to f8
				rook := state.Board.Squares[7][7]
				state.Board.Squares[5][7] = rook
				state.Board.Squares[7][7] = nil
			} else {
				// black queen-side: a8 to d8
				rook := state.Board.Squares[0][7]
				state.Board.Squares[3][7] = rook
				state.Board.Squares[0][7] = nil
			}
		}
	}

	// Handle promotion: use move.Promotion when provided, otherwise default to Queen
	if movedPiece != nil && movedPiece.Kind == Pawn {
		if movedPiece.Color == White && move.To.Rank == 7 {
			if move.Promotion != 0 {
				movedPiece.Kind = move.Promotion
			} else {
				movedPiece.Kind = Queen
			}
		}
		if movedPiece.Color == Black && move.To.Rank == 0 {
			if move.Promotion != 0 {
				movedPiece.Kind = move.Promotion
			} else {
				movedPiece.Kind = Queen
			}
		}
	}

	// 2. Detect check
	givesCheck := moveGivesCheck(state, move)

	// 3. Decide swap
	if givesCheck {
		state.SuppressNextSwap = true
	} else if state.SuppressNextSwap {
		state.SuppressNextSwap = false
	} else {
		applySwap(state, move.To)
	}

	// update en-passant target: only valid immediately after a pawn double-move
	state.HasEnPassant = false
	if movedPiece != nil && movedPiece.Kind == Pawn {
		if movedPiece.Color == White && move.From.Rank == 1 && move.To.Rank == 3 {
			state.HasEnPassant = true
			state.EnPassant = Position{File: move.From.File, Rank: 2}
		}
		if movedPiece.Color == Black && move.From.Rank == 6 && move.To.Rank == 4 {
			state.HasEnPassant = true
			state.EnPassant = Position{File: move.From.File, Rank: 5}
		}
	}

	// update castling rights: moving king revokes both, moving/capturing rooks revokes corresponding side
	if movedPiece != nil {
		if movedPiece.Kind == King {
			if movedPiece.Color == White {
				state.WhiteCanCastleKingSide = false
				state.WhiteCanCastleQueenSide = false
			} else {
				state.BlackCanCastleKingSide = false
				state.BlackCanCastleQueenSide = false
			}
		}
		if movedPiece.Kind == Rook {
			// if white rook moved from initial squares
			if movedPiece.Color == White {
				if move.From.File == 0 && move.From.Rank == 0 {
					state.WhiteCanCastleQueenSide = false
				}
				if move.From.File == 7 && move.From.Rank == 0 {
					state.WhiteCanCastleKingSide = false
				}
			} else {
				if move.From.File == 0 && move.From.Rank == 7 {
					state.BlackCanCastleQueenSide = false
				}
				if move.From.File == 7 && move.From.Rank == 7 {
					state.BlackCanCastleKingSide = false
				}
			}
		}
	}

	// if a rook was captured on its original square, revoke opponent rights
	if dest != nil && dest.Kind == Rook {
		if dest.Color == White {
			if move.To.File == 0 && move.To.Rank == 0 {
				state.WhiteCanCastleQueenSide = false
			}
			if move.To.File == 7 && move.To.Rank == 0 {
				state.WhiteCanCastleKingSide = false
			}
		} else {
			if move.To.File == 0 && move.To.Rank == 7 {
				state.BlackCanCastleQueenSide = false
			}
			if move.To.File == 7 && move.To.Rank == 7 {
				state.BlackCanCastleKingSide = false
			}
		}
	}

	// 4. Switch turn
	state.Turn = opposite(state.Turn)

	return nil
}

func moveGivesCheck(state *GameState, move Move) bool {
	_ = move
	// After the move is applied (caller ensures this), check whether any piece
	// belonging to the side who moved attacks the opponent king. This covers
	// direct attacks by the moved piece and discovered attacks uncovered by the move.

	mover := state.Turn
	opp := opposite(mover)

	// find opponent king
	var kingPos Position
	found := false
	for f := 0; f < 8; f++ {
		for r := 0; r < 8; r++ {
			p := state.Board.Squares[f][r]
			if p != nil && p.Kind == King && p.Color == opp {
				kingPos = Position{File: f, Rank: r}
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		return false
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

	// check if a piece at pos attacks target (king)
	attacks := func(pos, target Position, piece *Piece) bool {
		df := target.File - pos.File
		dr := target.Rank - pos.Rank
		adf := abs(df)
		adr := abs(dr)

		switch piece.Kind {
		case Rook:
			if (pos.File == target.File || pos.Rank == target.Rank) && pathClearBetween(pos, target) {
				return true
			}
		case Bishop:
			if adf == adr && pathClearBetween(pos, target) {
				return true
			}
		case Queen:
			if (pos.File == target.File || pos.Rank == target.Rank || adf == adr) && pathClearBetween(pos, target) {
				return true
			}
		case Knight:
			if (adf == 2 && adr == 1) || (adf == 1 && adr == 2) {
				return true
			}
		case Pawn:
			if piece.Color == White {
				if adf == 1 && dr == 1 {
					return true
				}
			} else {
				if adf == 1 && dr == -1 {
					return true
				}
			}
		case King:
			if adf <= 1 && adr <= 1 {
				return true
			}
		}
		return false
	}

	// scan all mover pieces to see if any attack the king
	for f := 0; f < 8; f++ {
		for r := 0; r < 8; r++ {
			p := state.Board.Squares[f][r]
			if p == nil || p.Color != mover {
				continue
			}
			if attacks(Position{File: f, Rank: r}, kingPos, p) {
				return true
			}
		}
	}

	return false
}

// squareAttacked reports whether `pos` is attacked by any piece of color `by`.
func squareAttacked(state *GameState, pos Position, by Color) bool {
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

	attacks := func(from Position, piece *Piece) bool {
		df := pos.File - from.File
		dr := pos.Rank - from.Rank
		adf := abs(df)
		adr := abs(dr)
		switch piece.Kind {
		case Rook:
			if (from.File == pos.File || from.Rank == pos.Rank) && pathClearBetween(from, pos) {
				return true
			}
		case Bishop:
			if adf == adr && pathClearBetween(from, pos) {
				return true
			}
		case Queen:
			if (from.File == pos.File || from.Rank == pos.Rank || adf == adr) && pathClearBetween(from, pos) {
				return true
			}
		case Knight:
			if (adf == 2 && adr == 1) || (adf == 1 && adr == 2) {
				return true
			}
		case Pawn:
			if piece.Color == White {
				if adf == 1 && dr == 1 {
					return true
				}
			} else {
				if adf == 1 && dr == -1 {
					return true
				}
			}
		case King:
			if adf <= 1 && adr <= 1 {
				return true
			}
		}
		return false
	}

	for f := 0; f < 8; f++ {
		for r := 0; r < 8; r++ {
			p := state.Board.Squares[f][r]
			if p == nil || p.Color != by {
				continue
			}
			if attacks(Position{File: f, Rank: r}, p) {
				return true
			}
		}
	}
	return false
}

func opposite(c Color) Color {
	if c == White {
		return Black
	}
	return White
}
