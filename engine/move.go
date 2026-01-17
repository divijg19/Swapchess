package engine

type Move struct {
	From Position
	To   Position
	// Promotion, if non-zero, indicates the piece kind to promote a pawn to.
	// If zero, the engine defaults to queen for promotions.
	Promotion PieceKind
}
