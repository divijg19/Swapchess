package view

import "github.com/divijg19/Swapchess/engine"

type GameStatus string

const (
	StatusInPlay    GameStatus = "in_play"
	StatusCheck     GameStatus = "check"
	StatusCheckmate GameStatus = "checkmate"
	StatusStalemate GameStatus = "stalemate"
)

func (s GameStatus) String() string {
	switch s {
	case StatusCheck:
		return "check"
	case StatusCheckmate:
		return "checkmate"
	case StatusStalemate:
		return "stalemate"
	default:
		return "in play"
	}
}

type CastlingRights struct {
	WhiteKingSide  bool
	WhiteQueenSide bool
	BlackKingSide  bool
	BlackQueenSide bool
}

type ViewPiece struct {
	Kind  engine.PieceKind
	Color engine.Color
	X     int
	Y     int
}

type ViewSquare struct {
	Occupied bool
	Kind     engine.PieceKind
	Color    engine.Color
}

type ViewState struct {
	Board            [8][8]ViewSquare
	Pieces           []ViewPiece
	Turn             engine.Color
	Status           GameStatus
	SuppressNextSwap bool
	HasEnPassant     bool
	EnPassant        engine.Position
	CastlingRights   CastlingRights
	LastMove         *engine.Move
	SwapEvent        *SwapEvent
}

type SwapEvent struct {
	A engine.Position
	B engine.Position
}

type SnapshotMeta struct {
	LastMove  *engine.Move
	SwapEvent *SwapEvent
}

// ViewStateFromGameState converts an engine.GameState into a render-agnostic ViewState.
func ViewStateFromGameState(s *engine.GameState) ViewState {
	return ViewStateFromGameStateWithMeta(s, SnapshotMeta{})
}

func ViewStateFromGameStateWithMeta(s *engine.GameState, meta SnapshotMeta) ViewState {
	vs := ViewState{
		Turn:             s.Turn,
		SuppressNextSwap: s.SuppressNextSwap,
		HasEnPassant:     s.HasEnPassant,
		EnPassant:        s.EnPassant,
		CastlingRights: CastlingRights{
			WhiteKingSide:  s.WhiteCanCastleKingSide,
			WhiteQueenSide: s.WhiteCanCastleQueenSide,
			BlackKingSide:  s.BlackCanCastleKingSide,
			BlackQueenSide: s.BlackCanCastleQueenSide,
		},
	}

	switch {
	case engine.IsCheckmate(s):
		vs.Status = StatusCheckmate
	case engine.IsStalemate(s):
		vs.Status = StatusStalemate
	case engine.IsInCheck(s, s.Turn):
		vs.Status = StatusCheck
	default:
		vs.Status = StatusInPlay
	}

	if meta.LastMove != nil {
		mv := *meta.LastMove
		vs.LastMove = &mv
	}
	if meta.SwapEvent != nil {
		ev := *meta.SwapEvent
		vs.SwapEvent = &ev
	}

	for f := 0; f < 8; f++ {
		for r := 0; r < 8; r++ {
			p := s.Board.Squares[f][r]
			if p == nil {
				continue
			}
			vs.Board[r][f] = ViewSquare{
				Occupied: true,
				Kind:     p.Kind,
				Color:    p.Color,
			}
			vs.Pieces = append(vs.Pieces, ViewPiece{
				Kind:  p.Kind,
				Color: p.Color,
				X:     f,
				Y:     r,
			})
		}
	}

	return vs
}
