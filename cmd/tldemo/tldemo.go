package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rsdoiel/termlib"
)

func main() {
	appName := filepath.Base(os.Args[0])
	helpText, fmtHelp := termlib.DemoHelpText, termlib.FmtHelp
	version, releaseDate, releaseHash, licenseText := termlib.Version, termlib.ReleaseDate, termlib.ReleaseHash, termlib.LicenseText
	showHelp, showLicense, showVersion := false, false, false

	// Standard Options
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.BoolVar(&showLicense, "license", false, "display license")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.Parse()

	out := os.Stdout

	if showHelp {
		fmt.Fprintf(out, "%s\n", fmtHelp(helpText, appName, version, releaseDate, releaseHash))
		os.Exit(0)
	}
	if showVersion {
		fmt.Fprintf(out, "%s %s %s\n", appName, version, releaseHash)
		os.Exit(0)
	}
	if showLicense {
		fmt.Fprintf(out, "%s\n", licenseText)
		os.Exit(0)
	}

	term := termlib.New()
	term.Clear()

	// Simulate a task with 10 steps
	totalSteps := 10
	for i := 0; i < totalSteps; i++ {
		// Print content in the upper part of the screen
		term.Move(1, 1)
		term.Print(fmt.Sprintf("Processing item %d of %d...", i+1, totalSteps))

		// Simulate work
		time.Sleep(500 * time.Millisecond)

		// Update progress counter at the bottom
		term.Move(term.GetTerminalHeight(), 1)
		term.ClrToEOL()
		progress := float64(i+1) / float64(totalSteps) * 100
		term.Print(fmt.Sprintf("Progress: %d%%", int(progress)))
		term.Refresh()
	}

	// Final message
	term.Move(term.GetTerminalHeight(), 1)
	term.ClrToEOL()
	term.Print("Task completed!")
	term.Refresh()
}
