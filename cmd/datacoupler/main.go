package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellbauer/data-coupler/internal/audit"
	"github.com/mitchellbauer/data-coupler/internal/config"
	"github.com/mitchellbauer/data-coupler/internal/credentials"
	"github.com/mitchellbauer/data-coupler/internal/engine"
	"github.com/mitchellbauer/data-coupler/internal/ui"

	// Self-registering packages — imported for side effects only.
	_ "github.com/mitchellbauer/data-coupler/internal/connector/csv"
	_ "github.com/mitchellbauer/data-coupler/internal/connector/mssql"
	_ "github.com/mitchellbauer/data-coupler/internal/connector/mysql"
	_ "github.com/mitchellbauer/data-coupler/internal/connector/odbc"
	_ "github.com/mitchellbauer/data-coupler/internal/connector/postgres"
	_ "github.com/mitchellbauer/data-coupler/internal/connector/sqlite"
	_ "github.com/mitchellbauer/data-coupler/internal/transform"
)

func main() {
	// Define flags.
	inPath := flag.String("in", "", "Path to source CSV file")
	outPath := flag.String("out", "", "Path to destination CSV file")
	profilePath := flag.String("profile", "", "Path to JSON profile")
	dryRun := flag.Bool("dry-run", false, "Validate inputs without writing file")

	flag.Parse()

	// Mode selection: no flags → GUI (LaunchApp handles its own initialization).
	if *inPath == "" && *profilePath == "" {
		ui.LaunchApp()
		return
	}

	// CLI mode validation.
	if *inPath == "" || *outPath == "" || *profilePath == "" {
		fmt.Println("❌ Error: CLI mode requires -in, -out, and -profile flags.")
		fmt.Println("Usage: data-coupler -in <file> -out <file> -profile <json>")
		os.Exit(1)
	}

	// CLI-only initialization (GUI path handles its own setup in LaunchApp).
	appDir := filepath.Dir(os.Args[0])
	credStore := credentials.NewFileStore(appDir)
	audit.SetLogPath(appDir)
	_ = audit.TrimLog(1000)
	settingsPath := filepath.Join(appDir, "settings.json")
	settings, _ := config.LoadSettings(settingsPath)

	// Run conversion.
	if err := runCLI(*inPath, *outPath, *profilePath, *dryRun, credStore); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Conversion Complete!")

	// Persist last-used profile path.
	settings.LastConnector = "csv"
	settings.LastProfilePath = *profilePath
	_ = config.SaveSettings(settingsPath, settings)
}

func runCLI(in, out, prof string, dry bool, creds credentials.Store) error {
	fmt.Printf("📂 Loading Profile: %s...\n", prof)
	profile, err := config.LoadProfile(prof)
	if err != nil {
		return err
	}

	// Override file paths from CLI flags (CSV profiles only).
	if in != "" {
		profile.Input.Path = in
	}
	if out != "" {
		profile.Output.Path = out
	}

	if dry {
		fmt.Println("running in dry-run mode... (Validation passed)")
		fmt.Printf("Would map %s -> %s using profile '%s'\n", in, out, profile.Name)
		return nil
	}

	fmt.Printf("⚙️  Running conversion using profile '%s'...\n", profile.Name)
	startTime := time.Now()
	count, runErr := engine.Run(profile, creds)

	errStr := ""
	if runErr != nil {
		errStr = runErr.Error()
	}
	_ = audit.AppendEntry(audit.AuditEntry{
		Timestamp:   startTime,
		ProfileName: profile.Name,
		ProfileID:   profile.ID,
		InputSource: "csv: " + filepath.Base(in),
		OutputPath:  out,
		RowsOut:     count,
		DurationMs:  time.Since(startTime).Milliseconds(),
		Error:       errStr,
	})

	if runErr != nil {
		return runErr
	}
	fmt.Printf("✅  %d rows processed.\n", count)
	return nil
}
