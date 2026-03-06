package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/divijg19/Swapchess/internal/app"
	cliui "github.com/divijg19/Swapchess/internal/ui/cli"
	tuiui "github.com/divijg19/Swapchess/internal/ui/tui"
)

type runFunc func(string) error

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, cliui.Run, tuiui.Run))
}

func run(args []string, stdout, stderr io.Writer, cliRunner, tuiRunner runFunc) int {
	flags := flag.NewFlagSet("swapchess", flag.ContinueOnError)
	flags.SetOutput(stderr)

	useCLI := flags.Bool("cli", false, "run CLI mode")
	mode := flags.String("mode", string(app.ModeTUI), "run mode: tui or cli")
	debugRenderer := flags.String("debug-renderer", "", "")
	flags.Usage = func() {
		fmt.Fprintf(stdout, "Usage: swapchess [--cli] [--mode=tui|cli]\n")
		fmt.Fprintf(stdout, "Default mode is the alt-screen terminal UI.\n")
	}

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	resolvedMode := *mode
	if *useCLI {
		resolvedMode = string(app.ModeCLI)
	}

	if resolvedMode != string(app.ModeCLI) && resolvedMode != string(app.ModeTUI) {
		fmt.Fprintf(stderr, "invalid mode %q; expected tui or cli\n", resolvedMode)
		return 2
	}

	if *debugRenderer != "" && *debugRenderer != string(app.RendererView) && *debugRenderer != string(app.RendererEngine) {
		fmt.Fprintf(stderr, "invalid debug renderer %q; expected view or engine\n", *debugRenderer)
		return 2
	}

	var err error
	switch app.Mode(resolvedMode) {
	case app.ModeCLI:
		err = cliRunner(*debugRenderer)
	default:
		err = tuiRunner(*debugRenderer)
	}

	if err != nil {
		fmt.Fprintf(stderr, "Error starting SwapChess: %v\n", err)
		return 1
	}

	return 0
}
