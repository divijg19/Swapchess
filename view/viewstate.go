package view

import "github.com/divijg19/Swapchess/engine"

type ViewPiece struct {
	Kind  engine.PieceKind
	Color engine.Color
	X     int
	Y     int
}

type ViewState struct {
	Pieces    []ViewPiece
	Turn      engine.Color
	SwapEvent *SwapEvent
}

type SwapEvent struct {
	A engine.Position
	B engine.Position
}

// ViewStateFromGameState converts an engine.GameState into a render-agnostic ViewState.
func ViewStateFromGameState(s *engine.GameState) ViewState {
	vs := ViewState{Turn: s.Turn}

	for f := 0; f < 8; f++ {
		for r := 0; r < 8; r++ {
			p := s.Board.Squares[f][r]
			if p == nil {
				continue
			}
			vp := ViewPiece{
				Kind:  p.Kind,
				Color: p.Color,
				X:     f,
				Y:     r,
			}
			vs.Pieces = append(vs.Pieces, vp)
		}
	}

	return vs
}
