package view

import (
	"testing"

	"github.com/divijg19/Swapchess/engine"
)

func TestViewStateFromGameStateWithMeta(t *testing.T) {
	state := engine.NewGame()
	state.SuppressNextSwap = true
	state.HasEnPassant = true
	state.EnPassant = engine.Position{File: 4, Rank: 2}
	state.WhiteCanCastleKingSide = false

	lastMove := &engine.Move{
		From: engine.Position{File: 4, Rank: 1},
		To:   engine.Position{File: 4, Rank: 3},
	}
	swapEvent := &SwapEvent{
		A: engine.Position{File: 4, Rank: 3},
		B: engine.Position{File: 6, Rank: 0},
	}

	vs := ViewStateFromGameStateWithMeta(state, SnapshotMeta{
		LastMove:  lastMove,
		SwapEvent: swapEvent,
	})

	if vs.Status != StatusInPlay {
		t.Fatalf("expected in-play status, got %s", vs.Status)
	}
	if !vs.SuppressNextSwap {
		t.Fatalf("expected suppress-next-swap to be true")
	}
	if !vs.HasEnPassant || vs.EnPassant != state.EnPassant {
		t.Fatalf("expected en-passant square to be propagated")
	}
	if vs.CastlingRights.WhiteKingSide {
		t.Fatalf("expected white king-side castling right to be false")
	}
	if vs.LastMove == nil || *vs.LastMove != *lastMove {
		t.Fatalf("expected last move metadata to be copied")
	}
	if vs.SwapEvent == nil || *vs.SwapEvent != *swapEvent {
		t.Fatalf("expected swap event metadata to be copied")
	}

	e1 := vs.Board[0][4]
	if !e1.Occupied || e1.Kind != engine.King || e1.Color != engine.White {
		t.Fatalf("expected white king on e1 in board projection, got %+v", e1)
	}

	var whitePawns int
	for _, piece := range vs.Pieces {
		if piece.Kind == engine.Pawn && piece.Color == engine.White {
			whitePawns++
		}
	}
	if whitePawns != 8 {
		t.Fatalf("expected 8 white pawns, got %d", whitePawns)
	}
}

func TestViewStateStatusDetectsCheck(t *testing.T) {
	state := &engine.GameState{Turn: engine.Black}
	state.Board.Squares[4][0] = &engine.Piece{Kind: engine.King, Color: engine.White}
	state.Board.Squares[4][7] = &engine.Piece{Kind: engine.King, Color: engine.Black}
	state.Board.Squares[4][5] = &engine.Piece{Kind: engine.Rook, Color: engine.White}

	vs := ViewStateFromGameState(state)
	if vs.Status != StatusCheck {
		t.Fatalf("expected check status, got %s", vs.Status)
	}
}
