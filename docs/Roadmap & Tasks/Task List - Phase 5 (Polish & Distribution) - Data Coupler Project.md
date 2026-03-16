---
date created: Sunday, March 15th 2026, 12:00:00 pm
date modified: Sunday, March 15th 2026, 12:00:00 pm
---

**Focus:** Audit Logging, Auto-Update, Code Signing, & Installer Packaging
**Parent:** [[Roadmap - Data Coupler Project]]
**Prerequisite:** [[Task List - Phase 4 (Fishbowl Templates) - Data Coupler Project]]

> **Why a polish phase?**
> By the end of Phase 4, the app is fully functional. Phase 5 is about trustworthiness and distribution: keeping a record of what ran (audit log), making sure users know when a better version exists (auto-update), and eliminating OS security warnings on first run (code signing + packaging). These aren't features — they're the difference between a developer tool and software someone's willing to hand to a non-technical colleague.

---

# 📋 Stage 1: Audit Log

*Every conversion run — successful or failed — leaves a timestamped record. The log is a plain newline-delimited JSON file, human-readable with a text editor, and requires no GUI to view.*

## Step 1.1: Log Entry Definition

* [ ] Define `AuditEntry` struct in `internal/audit/audit.go`:
  ```go
  type AuditEntry struct {
      Timestamp   time.Time `json:"timestamp"`
      ProfileName string    `json:"profileName"`
      ProfileID   string    `json:"profileId"`
      InputSource string    `json:"inputSource"` // e.g., "mssql: macola-prod / IM_ITEM query"
      OutputPath  string    `json:"outputPath"`
      RowsOut     int       `json:"rowsOut"`
      DurationMs  int64     `json:"durationMs"`
      Error       string    `json:"error,omitempty"` // empty on success
  }
  ```
* [ ] Log file path: `audit.log` in the application's data directory (same folder as `credentials.bin`).
* [ ] Format: one JSON object per line (newline-delimited JSON / NDJSON). Human-readable with a text editor.

## Step 1.2: Write & Trim Logic

* [ ] `AppendEntry(entry AuditEntry) error`:
  * Marshal entry to JSON, append a line to `audit.log`.
  * ⚠️ **Constraint:** Open the file in append mode — never rewrite the whole file on each append.
* [ ] `TrimLog(maxEntries int) error`:
  * Called on application startup.
  * Reads `audit.log`, keeps only the last `maxEntries` lines, rewrites.
  * Default `maxEntries`: 1000.
  * If the file has fewer than `maxEntries` lines, do nothing.
* [ ] Unit tests:
  * `AppendEntry` writes a valid JSON line.
  * `TrimLog` correctly reduces a 1200-line file to 1000 lines (keeps the newest entries).
  * `TrimLog` on a missing file does nothing (no error).

## Step 1.3: Wire into the Engine

* [ ] In `internal/engine/engine.go`, after `Run()` completes (success or failure), call `audit.AppendEntry()`.
* [ ] Pass the audit package's store as a dependency (or use a package-level function — keep it simple).
* [ ] Call `audit.TrimLog(1000)` once during application startup in `main.go`.
* [ ] **Milestone:** Run two conversions (one success, one intentional failure with a bad input file). Open `audit.log` and verify both entries are present with correct timestamps, row counts, and error fields.

---

# 🔔 Stage 2: Auto-Update Check

*The app checks for a newer GitHub release on startup and shows a non-blocking banner if one is found. It never downloads or installs anything automatically.*

## Step 2.1: Version Constant

* [ ] Define `AppVersion = "0.4.0"` (or current version) as a constant in `cmd/datacoupler/main.go` or a dedicated `internal/version/version.go`.
* [ ] Use Go build-time injection as an optional enhancement: `-ldflags "-X main.AppVersion=0.4.0"` in `build.bat`.

## Step 2.2: Update Check Logic

* [ ] Create `internal/updater/updater.go`.
* [ ] `CheckLatestRelease(repoOwner, repoName string) (string, error)`:
  * Fetches `https://api.github.com/repos/<owner>/<repo>/releases/latest`.
  * Parses the `"tag_name"` field (e.g., `"v0.5.0"`).
  * Returns the version string.
  * ⚠️ **Constraint:** Set a 5-second HTTP timeout. On any error (no internet, rate limit, parse failure), return `("", err)` silently — never surface this error to the user.
* [ ] `IsNewer(current, latest string) bool`:
  * Strips leading `v` from both strings.
  * Compares semver (major.minor.patch). Use `golang.org/x/mod/semver` or simple string splitting — no heavy semver library needed.
* [ ] Unit tests:
  * `IsNewer("0.3.0", "0.4.0")` → `true`
  * `IsNewer("0.4.0", "0.4.0")` → `false`
  * `IsNewer("0.5.0", "0.4.0")` → `false`
  * `IsNewer("0.3.0", "v0.4.0")` → `true` (leading `v` stripped)

## Step 2.3: Home Screen Banner

* [ ] On app startup, launch a goroutine that calls `CheckLatestRelease()`.
* [ ] If `IsNewer(AppVersion, latest)` is true, post an update to the UI thread:
  * Show a dismissible info banner on the home screen: "Version X.Y.Z is available. Click to open the download page."
  * Clicking the banner opens the GitHub releases URL in the system browser (`fyne.App.OpenURL()`).
  * A dismiss (×) button hides the banner for the rest of the session (not permanently).
* [ ] If the check fails or returns no newer version, no UI change.
* [ ] ⚠️ **Constraint:** The banner must appear via the UI event loop — never call Fyne widget methods directly from a goroutine.
* [ ] **Milestone:** Temporarily set `AppVersion` to `"0.1.0"` and verify the banner appears with the correct latest version string.

---

# 🔏 Stage 3: Code Signing

*Unsigned binaries trigger "Windows protected your PC" and macOS Gatekeeper warnings. Code signing removes these for end users.*

## Step 3.1: Windows (Authenticode)

* [ ] Obtain a code signing certificate (EV or OV from a CA like DigiCert, Sectigo, or use a self-signed cert for internal distribution).
* [ ] Document the signing command in `docs/Technical Design and Setup/Maintenance and Build Procedure - Data Coupler Project.md`:
  ```bat
  signtool sign /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 /f certificate.pfx /p <password> DataCoupler.exe
  ```
* [ ] Update `build.bat` with an optional signing step gated on the presence of `certificate.pfx`:
  ```bat
  IF EXIST certificate.pfx (
      signtool sign ...
  ) ELSE (
      echo [SKIP] Code signing certificate not found. Binary will not be signed.
  )
  ```
* [ ] Test: right-click the signed `.exe` → Properties → Digital Signatures tab shows the certificate.

## Step 3.2: macOS (Developer ID)

* [ ] Obtain a Developer ID Application certificate from the Apple Developer portal.
* [ ] Document the signing and notarization commands in `Maintenance and Build Procedure`:
  ```bash
  codesign --deep --force --options runtime \
      --sign "Developer ID Application: <Name> (<TeamID>)" \
      DataCoupler.app

  xcrun notarytool submit DataCoupler.dmg \
      --apple-id <email> --team-id <TeamID> --password <app-specific-password> \
      --wait

  xcrun stapler staple DataCoupler.dmg
  ```
* [ ] Update the macOS build script with optional signing step (gated on cert presence).
* [ ] Test: verify Gatekeeper passes (`spctl --assess --verbose DataCoupler.app`).

---

# 📦 Stage 4: Installer & Packaging

*The final deliverable is a single-click install experience on Windows, macOS, and Linux. No ZIP files, no "extract this folder somewhere."*

## Step 4.1: Windows MSI (WiX Toolset)

* [ ] Install WiX Toolset v4 (`dotnet tool install --global wix`).
* [ ] Create `installer/windows/DataCoupler.wxs` defining:
  * Product name, version, manufacturer, upgrade GUID.
  * Install directory: `%ProgramFiles%\DataCoupler\`.
  * Start Menu shortcut.
  * Desktop shortcut (optional, user-selectable).
  * Uninstaller entry in Add/Remove Programs.
* [ ] Add MSI build step to `build.bat`:
  ```bat
  wix build installer\windows\DataCoupler.wxs -o dist\DataCoupler-<version>.msi
  ```
* [ ] Sign the MSI with Authenticode (same cert as the EXE).
* [ ] Test: install on a clean Windows VM, verify app runs, verify uninstall removes all files cleanly.

## Step 4.2: macOS DMG

* [ ] Create `installer/macos/` folder.
* [ ] Build `.app` bundle:
  * `Info.plist` with bundle ID, version, display name.
  * `Icon.icns` (convert `Icon.png` using `iconutil`).
  * Signed binary inside `Contents/MacOS/`.
* [ ] Create DMG with a background image and a drag-to-`/Applications` layout:
  ```bash
  create-dmg \
      --volname "Data Coupler" \
      --window-size 600 400 \
      --icon-size 100 \
      --app-drop-link 450 200 \
      "dist/DataCoupler-<version>.dmg" \
      "DataCoupler.app"
  ```
* [ ] Notarize and staple the DMG (see Stage 3.2).
* [ ] Test: install on a clean macOS VM, verify Gatekeeper passes on first launch.

## Step 4.3: Linux AppImage

* [ ] Download `appimagetool` for the build machine.
* [ ] Create `installer/linux/AppDir/` with:
  * `AppRun` shell script (sets `$PATH` and launches the binary).
  * `DataCoupler.desktop` file.
  * `DataCoupler.png` icon.
  * The compiled binary.
* [ ] Build:
  ```bash
  appimagetool AppDir/ dist/DataCoupler-<version>-x86_64.AppImage
  ```
* [ ] Test: run the AppImage on Ubuntu (no install required), verify it launches.

## Step 4.4: Build Script Consolidation

* [ ] Update `build.bat` to produce all three platform targets (cross-compile using `GOOS` / `GOARCH`):
  ```bat
  SET VERSION=0.4.0
  go build -ldflags "-X main.AppVersion=%VERSION%" -o dist/DataCoupler-windows.exe ./cmd/datacoupler
  GOOS=darwin GOARCH=amd64 go build ... -o dist/DataCoupler-macos ./cmd/datacoupler
  GOOS=linux  GOARCH=amd64 go build ... -o dist/DataCoupler-linux  ./cmd/datacoupler
  ```
* [ ] ⚠️ **Constraint:** The Fyne CGO dependency may complicate cross-compilation. Document any platform-specific build requirements (e.g., macOS builds must run on macOS for CGO).
* [ ] **Final Milestone:** On each target OS, a fresh install from the packaged artifact (MSI / DMG / AppImage) results in a working application. No "Windows protected your PC" dialog on Windows. No Gatekeeper warning on macOS.

---

# ✅ Stage 5: Final Verification Checklist

* [ ] `go test ./...` passes with zero errors on all three platforms (or documented exceptions for platform-specific packages).
* [ ] Audit log: run 5 conversions, verify 5 entries in `audit.log` with correct data.
* [ ] Auto-update: test banner appears with a mock older `AppVersion`.
* [ ] Code signing: verify signed EXE on Windows and signed/notarized `.app` on macOS.
* [ ] Installer: test clean install and clean uninstall on each platform.
* [ ] Check the Future Backlog in the Roadmap and decide if any items have become high enough priority to pull into a Phase 6.
