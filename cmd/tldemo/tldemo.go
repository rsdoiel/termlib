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

	term := termlib.New(out)
	term.Clear()

	// Simulate a task with 10 steps
	totalSteps := 10
	demoShown := false

	for i := 0; i < totalSteps; i++ {
		// Print content in the upper part of the screen
		term.Move(1, 1)
		term.Print(fmt.Sprintf("Processing item %d of %d...", i+1, totalSteps))
		term.Printf("\nWidth: %d\n", term.GetTerminalWidth())
		term.Printf("Height: %d\n", term.GetTerminalHeight())
		row, col := term.GetCurPos()
		term.Printf("row %d, col: %d\n\n", row, col)

		// Show color and style demo after the first few steps
		if i == 3 && !demoShown {
			showStyleDemo(term)
			demoShown = true
			time.Sleep(2 * time.Second) // Give time to view the demo
		}

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

func showStyleDemo(term *termlib.Terminal) {
	// Clear the middle section for our demo
	for row := 3; row < term.GetTerminalHeight()-2; row++ {
		term.Move(row, 1)
		term.ClrToEOL()
	}

	// Title
	term.Move(3, 1)
	term.SetBold()
	term.Print("=== Terminal Color & Style Demo ===")
	term.ResetStyle()

	// Foreground colors section
	term.Move(5, 1)
	term.SetBold()
	term.Print("Foreground Colors:")
	term.ResetStyle()

	term.Move(6, 3)
	term.SetFgColor(termlib.Black)
	term.Print("■ Black")

	term.Move(7, 3)
	term.SetFgColor(termlib.Red)
	term.Print("■ Red")

	term.Move(8, 3)
	term.SetFgColor(termlib.Green)
	term.Print("■ Green")

	term.Move(9, 3)
	term.SetFgColor(termlib.Yellow)
	term.Print("■ Yellow")

	term.Move(10, 3)
	term.SetFgColor(termlib.Blue)
	term.Print("■ Blue")

	term.Move(11, 3)
	term.SetFgColor(termlib.Magenta)
	term.Print("■ Magenta")

	term.Move(12, 3)
	term.SetFgColor(termlib.Cyan)
	term.Print("■ Cyan")

	term.Move(13, 3)
	term.SetFgColor(termlib.White)
	term.Print("■ White")

	// Background colors section
	term.Move(5, 25)
	term.SetBold()
	term.Print("Background Colors:")
	term.ResetStyle()

	term.Move(6, 27)
	term.SetBgColor(termlib.BlackBg)
	term.Print(" Black  ")

	term.Move(7, 27)
	term.SetBgColor(termlib.RedBg)
	term.Print(" Red    ")

	term.Move(8, 27)
	term.SetBgColor(termlib.GreenBg)
	term.Print(" Green  ")

	term.Move(9, 27)
	term.SetBgColor(termlib.YellowBg)
	term.Print(" Yellow ")

	term.Move(10, 27)
	term.SetBgColor(termlib.BlueBg)
	term.Print(" Blue   ")

	term.Move(11, 27)
	term.SetBgColor(termlib.MagentaBg)
	term.Print(" Magenta")

	term.Move(12, 27)
	term.SetBgColor(termlib.CyanBg)
	term.Print(" Cyan   ")

	term.Move(13, 27)
	term.SetBgColor(termlib.WhiteBg)
	term.Print(" White  ")

	// Text styles section
	term.Move(5, 50)
	term.SetBold()
	term.Print("Text Styles:")
	term.ResetStyle()

	term.Move(6, 52)
	term.SetBold()
	term.Print("Bold text")

	term.Move(7, 52)
	term.SetItalic()
	term.Print("Italic text")

	term.Move(8, 52)
	term.SetBold()
	term.SetItalic()
	term.Print("Bold Italic")

	// Combined styles section
	term.Move(10, 50)
	term.SetBold()
	term.Print("Combined Styles:")
	term.ResetStyle()

	term.Move(11, 52)
	term.SetFgColor(termlib.White)
	term.SetBgColor(termlib.BlueBg)
	term.Print(" White on Blue ")

	term.Move(12, 52)
	term.SetFgColor(termlib.Black)
	term.SetBgColor(termlib.YellowBg)
	term.Print(" Black on Yellow ")

	term.Move(13, 52)
	term.SetFgColor(termlib.Red)
	term.SetBgColor(termlib.WhiteBg)
	term.SetBold()
	term.Print(" Red Bold on White ")

	// Reset all styles
	term.ResetStyle()

	// Instructions
	term.Move(term.GetTerminalHeight()-3, 1)
	term.ClrToEOL()
	term.Print("Color and style demo shown during processing...")
}

