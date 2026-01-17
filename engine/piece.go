package engine

type Color int

const (
	White Color = iota
	Black
)

type PieceKind int

const (
	Pawn PieceKind = iota
	Knight
	Bishop
	Rook
	Queen
	King
)

type Piece struct {
	Kind  PieceKind
	Color Color
}

func (c Color) String() string {
	switch c {
	case White:
		return "White"
	case Black:
		return "Black"
	default:
		return "Unknown"
	}
}

func (k PieceKind) String() string {
	switch k {
	case Pawn:
		return "Pawn"
	case Knight:
		return "Knight"
	case Bishop:
		return "Bishop"
	case Rook:
		return "Rook"
	case Queen:
		return "Queen"
	case King:
		return "King"
	default:
		return "Unknown"
	}
}
