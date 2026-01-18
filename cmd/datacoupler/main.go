package main

import (
	"data-coupler/internal/config"
	"data-coupler/internal/engine"
	"data-coupler/internal/ui"
	"flag"
	"fmt"
	"os"
)

func main() {
	// 1. Define Flags
	inPath := flag.String("in", "", "Path to source CSV file")
	outPath := flag.String("out", "", "Path to destination CSV file")
	profilePath := flag.String("profile", "", "Path to JSON profile")
	dryRun := flag.Bool("dry-run", false, "Validate inputs without writing file")

	flag.Parse()

	// 2. Mode Selection
	// If no inputs are provided, assume the user wants the GUI.
	if *inPath == "" && *profilePath == "" {
		runGUI()
		return
	}

	// 3. CLI Mode Validation
	if *inPath == "" || *outPath == "" || *profilePath == "" {
		fmt.Println("❌ Error: CLI mode requires -in, -out, and -profile flags.")
		fmt.Println("Usage: data-coupler -in <file> -out <file> -profile <json>")
		os.Exit(1)
	}

	// 4. Run Conversion
	if err := runCLI(*inPath, *outPath, *profilePath, *dryRun); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Conversion Complete!")
}

func runCLI(in, out, prof string, dry bool) error {
	// A. Load Profile
	fmt.Printf("📂 Loading Profile: %s...\n", prof)
	profile, err := config.LoadProfile(prof)
	if err != nil {
		return err
	}

	// B. Dry Run Check
	if dry {
		fmt.Println("running in dry-run mode... (Validation passed)")
		fmt.Printf("Would map %s -> %s using profile '%s'\n", in, out, profile.Name)
		return nil
	}

	// C. Execute Engine
	fmt.Printf("⚙️  Mapping %s -> %s...\n", in, out)
	return engine.Run(in, out, profile)
}

func runGUI() {
	ui.LaunchApp()
}
