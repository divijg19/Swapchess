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

**Released:** `v1.0.0`

Completed milestones:

1. Pure Go engine (rules, board state, swap logic)
2. Terminal UI (TUI) using **Bubble Tea**
3. CLI fallback mode for terminal-first play and debugging
4. Packaging and native release binaries as build artifacts

---

## Architecture

`Swapchess` is built with a strict separation between game logic and rendering.

```
engine/   → Pure chess + `Swapchess` rules (no UI dependencies)
view/     → Render-agnostic game snapshot mapping
internal/
  ├─ app/        → Shared terminal session/input state
  ├─ render/text → Shared text board/status renderers
  └─ ui/         → Terminal mode implementations
cmd/      → Public launchers
  ├─ swapchess/ → Canonical terminal launcher
  └─ gfx/       → Reserved native 2D renderer
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
* Keyboard-driven with board focus and command prompt
* Unicode piece rendering with file-based asset overrides from `assets/pieces`
* Shared input validation, move parsing, promotion flow, and undo with CLI mode
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

## Build & Run

Current terminal launcher:

```bash
go run ./cmd/swapchess
```

CLI mode:

```bash
go run ./cmd/swapchess --cli
```

Explicit mode selection:

```bash
go run ./cmd/swapchess --mode=tui
go run ./cmd/swapchess --mode=cli
```

The hidden debug renderer flag can be used for development comparisons:

```bash
go run ./cmd/swapchess --debug-renderer=view
go run ./cmd/swapchess --debug-renderer=engine
```

`Swapchess` is still intended to ship as a **single native executable** per platform with no external runtime dependencies.

Print the current release version:

```bash
go run ./cmd/swapchess --version
```

Canonical release builds are produced by **GitHub Actions** on pushed version tags like `v1.0.0`.

CI workflow:

* runs tests and build checks on pushes and pull requests

Release workflow:

* runs on tags matching `v*.*.*`
* builds cross-platform artifacts
* publishes them to the GitHub Release

For local reproduction of the release packaging step, you can still run:

```bash
./scripts/build_release_artifacts.sh
```

Local artifacts are written to:

```text
bin/releases/v1.0.0/
```

The release artifact build emits packaged downloads for:

* Linux `amd64`
* Linux `arm64`
* macOS `amd64`
* macOS `arm64`
* Windows `amd64`

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
