package engine

type Position struct {
	File int // 0–7
	Rank int // 0–7
}

type Board struct {
	Squares [8][8]*Piece
}
