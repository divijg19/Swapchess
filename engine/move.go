package engine

type Move struct {
	From Position
	To   Position
	// Promotion holds the requested promotion piece kind when this move is a pawn promotion.
	Promotion PieceKind
	// PromotionSet explicitly indicates that Promotion was intentionally provided.
	// This avoids ambiguity because PieceKind zero-value is Pawn.
	PromotionSet bool
}

// HasExplicitPromotion reports whether the move explicitly carries a promotion choice.
// It supports both the new explicit flag and older callers that only set Promotion.
func (m Move) HasExplicitPromotion() bool {
	return m.PromotionSet || m.Promotion != Pawn
}
