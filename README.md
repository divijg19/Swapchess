# `Swapchess`

*A small, native chess variant exploring controlled chaos in a deterministic system.*

---

## Overview

**`Swapchess`** is a chess variant implemented in **Go** where every legal move destabilizes the board in a subtle but strategic way.

> After every legal move, the moved piece swaps position with a random piece of the same color.

The result is a game that feels familiar at first, but quickly evolves into something unpredictable, tactical, and uniquely tense.

This project is intentionally scoped as a **complete, offline, native game**, not a chess replacement or an AI showcase.

---

## Core Rules

`Swapchess` follows standard chess rules with one additional mechanic.

### Swap Rule

* After **every legal move**, the moved piece swaps position with **one randomly chosen piece of the same color** owned by the player.

### Swap Suppression (Important)

* **No swap occurs if the move delivers check**
* **No swap occurs on the opponent’s immediately following move**
* All other legal moves trigger **exactly one swap**

These rules are fixed and define the game.

---

## Design Goals

* Preserve the structure and readability of classical chess
* Introduce instability through a *single*, well-defined rule
* Keep randomness controlled, explainable, and strategic
* Reward planning around checks as moments of positional stability

This is a game about **managing chaos**, not surrendering to it.

---

## Project Status

**In development:** January – February 2026

Planned milestones:

1. Pure Go engine (rules, board state, swap logic)
2. Terminal UI (TUI) using **Bubble Tea**
3. Native windowed 2D UI using **Ebiten (Ebitengine)**
4. Visual polish, packaging, and release binaries

---

## Architecture

`Swapchess` is built with a strict separation between game logic and rendering.

```
engine/   → Pure chess + `Swapchess` rules (no UI dependencies)
ui/       → Rendering & input layers
  ├─ tui/ → Bubble Tea terminal interface
  └─ gfx/ → Ebiten native 2D renderer
assets/   → Embedded piece & board art
```

### Key Principles

* The engine is fully deterministic and testable
* Rendering consumes a render-agnostic `ViewState`
* No UI layer mutates game state directly
* All randomness is seedable

---

## User Interfaces

### Terminal UI (TUI)

* Built with **Bubble Tea**
* Keyboard-driven
* Used for rule validation and fast iteration

### Native 2D UI

* Built with **Ebiten**
* Mouse-based drag & drop
* Clean, minimal chessboard visuals
* Subtle animations for swaps and highlights

There is **no browser or web-based UI**.

---

## Game Modes

* Local Human vs Human
* Local Human vs Simple Bot

The bot is intentionally minimal and exists to enable solo play.

---

## Non-Goals

The following are explicitly **out of scope**:

* Online multiplayer
* Rankings or matchmaking
* Monetization
* Advanced chess engines
* Additional variants
* Web or mobile versions

This project values *completion over expansion*.

---

## Build & Run (Planned)

Once implemented, `Swapchess` will be distributed as a **single native executable** per platform with no external runtime dependencies.

Target platforms:

* Linux (primary)
* macOS
* Windows

---

## Why This Exists

`Swapchess` exists as:

* A finished, rules-driven game
* A demonstration of clean engine/UI separation in Go
* An exploration of how a single rule can destabilize a perfect system without destroying it

It is designed to be small, intentional, and complete.

---

> *“`Swapchess` is a small, native game that explores how a single rule can destabilize a perfect system without destroying it.”*
