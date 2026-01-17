package engine

// NewGame returns a GameState set to a standard chess starting position.
func NewGame() *GameState {
	gs := &GameState{
		Turn:                    White,
		RandSeed:                1,
		HasEnPassant:            false,
		WhiteCanCastleKingSide:  true,
		WhiteCanCastleQueenSide: true,
		BlackCanCastleKingSide:  true,
		BlackCanCastleQueenSide: true,
	}

	// Place white pieces
	gs.Board.Squares[0][0] = &Piece{Kind: Rook, Color: White}
	gs.Board.Squares[1][0] = &Piece{Kind: Knight, Color: White}
	gs.Board.Squares[2][0] = &Piece{Kind: Bishop, Color: White}
	gs.Board.Squares[3][0] = &Piece{Kind: Queen, Color: White}
	gs.Board.Squares[4][0] = &Piece{Kind: King, Color: White}
	gs.Board.Squares[5][0] = &Piece{Kind: Bishop, Color: White}
	gs.Board.Squares[6][0] = &Piece{Kind: Knight, Color: White}
	gs.Board.Squares[7][0] = &Piece{Kind: Rook, Color: White}
	for f := 0; f < 8; f++ {
		gs.Board.Squares[f][1] = &Piece{Kind: Pawn, Color: White}
	}

	// Place black pieces
	gs.Board.Squares[0][7] = &Piece{Kind: Rook, Color: Black}
	gs.Board.Squares[1][7] = &Piece{Kind: Knight, Color: Black}
	gs.Board.Squares[2][7] = &Piece{Kind: Bishop, Color: Black}
	gs.Board.Squares[3][7] = &Piece{Kind: Queen, Color: Black}
	gs.Board.Squares[4][7] = &Piece{Kind: King, Color: Black}
	gs.Board.Squares[5][7] = &Piece{Kind: Bishop, Color: Black}
	gs.Board.Squares[6][7] = &Piece{Kind: Knight, Color: Black}
	gs.Board.Squares[7][7] = &Piece{Kind: Rook, Color: Black}
	for f := 0; f < 8; f++ {
		gs.Board.Squares[f][6] = &Piece{Kind: Pawn, Color: Black}
	}

	return gs
}

// Clone returns a deep copy of the GameState.
func (g *GameState) Clone() *GameState {
	ng := &GameState{
		Turn:                    g.Turn,
		RandSeed:                g.RandSeed,
		SuppressNextSwap:        g.SuppressNextSwap,
		HasEnPassant:            g.HasEnPassant,
		EnPassant:               g.EnPassant,
		WhiteCanCastleKingSide:  g.WhiteCanCastleKingSide,
		WhiteCanCastleQueenSide: g.WhiteCanCastleQueenSide,
		BlackCanCastleKingSide:  g.BlackCanCastleKingSide,
		BlackCanCastleQueenSide: g.BlackCanCastleQueenSide,
	}

	for f := 0; f < 8; f++ {
		for r := 0; r < 8; r++ {
			p := g.Board.Squares[f][r]
			if p != nil {
				cp := *p
				ng.Board.Squares[f][r] = &cp
			}
		}
	}

	return ng
}
