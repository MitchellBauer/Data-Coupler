---
date created: Saturday, January 17th 2026, 9:04:05 am
date modified: Wednesday, February 4th 2026, 6:10:00 pm
---

**Parent:** [[Project - Data Coupler Summary]]
**Last Updated:** 2026-01-17

# 1. Yearly Environment Update

*Perform these steps if you haven't touched the project in 6+ months.*

## A. Update Go (The Language)

1. **Check Current Version:** Open terminal and run `go version`.
2. **Download Update:** Go to [go.dev/dl](https://go.dev/dl/) and download the **Microsoft Windows** installer (`.msi`).
3. **Install:** Run the installer. It will automatically replace the old version and update your system `PATH`.
4. **Verify:** Open a *new* terminal window and run `go version` to confirm the update.

## B. Update Dependencies

1. **Open Terminal** in the project root.
2. **Update Fyne CLI:**

```powershell
go install fyne.io/fyne/v2/cmd/fyne@latest

```

3. **Update Libraries:**

```powershell
go get -u ./...
go mod tidy

```

*(Note: `go mod tidy` removes unused packages to keep the build clean.)*

---

# 2. Standard Build (`build.bat`)

The preferred build script is `build.bat` in the project root. Run it by double-clicking or from the terminal:

```bat
build.bat
```

It performs four steps in order:

| Step | Action |
|---|---|
| 1 | `go test ./...` — aborts on failure |
| 2 | Compiles `Data Coupler.exe` with version injected via `-ldflags` |
| 3 | Signs the EXE if `certificate.pfx` is present (skips silently otherwise) |
| 4 | Builds `dist\DataCoupler-<VERSION>.msi` if WiX is installed (skips silently otherwise) |

**Set the release version** by editing `SET VERSION=X.Y.Z` near the top of `build.bat`.

**Enable code signing** by placing `certificate.pfx` in the project root and setting `SIGN_PASSWORD`:

```bat
SET SIGN_PASSWORD=<pfx-password>
build.bat
```

**Enable MSI packaging** by installing WiX Toolset v4:

```bat
dotnet tool install --global wix
```

---

# 3. Code Signing

## Windows — Authenticode

Obtain an OV or EV code signing certificate from a CA (DigiCert, Sectigo, etc.). Export as PFX.

Manual signing command:

```bat
signtool sign /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 /f certificate.pfx /p <password> "Data Coupler.exe"
```

Verify:

```bat
signtool verify /pa "Data Coupler.exe"
```

Or right-click → Properties → Digital Signatures tab.

## macOS — Developer ID + Notarization

> macOS builds must be compiled on a macOS machine (CGO/Fyne requirement).

1. Obtain a **Developer ID Application** certificate from the Apple Developer portal.
2. Build on macOS:
   ```bash
   go build -ldflags="-X .../version.AppVersion=X.Y.Z" -o DataCoupler ./cmd/datacoupler
   ```
3. Sign the `.app` bundle:
   ```bash
   codesign --deep --force --options runtime \
       --sign "Developer ID Application: <Name> (<TeamID>)" \
       DataCoupler.app
   ```
4. Create DMG (requires `brew install create-dmg`):
   ```bash
   create-dmg --volname "Data Coupler" --window-size 600 400 \
       --icon-size 100 --app-drop-link 450 200 \
       DataCoupler-X.Y.Z.dmg DataCoupler.app
   ```
5. Notarize and staple:
   ```bash
   xcrun notarytool submit DataCoupler-X.Y.Z.dmg \
       --apple-id <email> --team-id <TeamID> --password <app-specific-password> --wait
   xcrun stapler staple DataCoupler-X.Y.Z.dmg
   ```
6. Verify:
   ```bash
   spctl --assess --verbose DataCoupler.app
   ```

---

# 4. Installer Packaging

## Windows MSI

Automatically built by `build.bat` when `wix` is on PATH.

WiX config: `installer/windows/DataCoupler.wxs`

- Installs to `%ProgramFiles%\DataCoupler\`
- Creates a Start Menu shortcut under `Data Coupler\`
- Adds an uninstaller entry in Add/Remove Programs

Manual build:

```bat
wix build installer\windows\DataCoupler.wxs -d Version=X.Y.Z -o dist\DataCoupler-X.Y.Z.msi
```

## Linux AppImage

> Must be built on a Linux machine.

1. Download `appimagetool` from https://github.com/AppImage/AppImageKit/releases.
2. Create `installer/linux/AppDir/` containing `AppRun`, `DataCoupler.desktop`, `DataCoupler.png`, and the binary.
3. Build:
   ```bash
   appimagetool AppDir/ dist/DataCoupler-X.Y.Z-x86_64.AppImage
   ```

---

# 5. Release Checklist

- [ ] Update `SET VERSION=X.Y.Z` in `build.bat`
- [ ] Update `var AppVersion = "X.Y.Z"` in `internal/version/version.go`
- [ ] Run `go test ./...` — all tests pass
- [ ] Run `build.bat` — EXE, signed EXE, and MSI produced
- [ ] Smoke-test on a clean Windows VM (install from MSI, verify no SmartScreen warning if signed)
- [ ] Create a GitHub Release tagged `vX.Y.Z` and attach the MSI
- [ ] Verify the auto-update banner appears in the previous installed version after the release is published

---

# 6. Legacy Build Process

*Do not build manually. Use the automation script to ensure all logic tests pass before creating the executable.*

## Step 1: Run the Script

Double-click `build_and_test.ps1` in the project root, or run it via PowerShell:

```powershell
.\build_and_test.ps1

```

## Step 2: Interpret Results

* **✅ Success:** The script will output "Success! Application built." You will find the new `.exe` in the root folder.
* **❌ Failure:** The script will stop immediately if **any** test fails. Read the red error output to see which logic (or GUI interaction) broke. Fix the code and re-run.

---

# 3. Troubleshooting

*Common issues after a long hiatus.*

* **Error: `gcc: executable file not found**`
* **Cause:** The C compiler link is broken (required by Fyne).
* **Fix:** Re-install **TDM-GCC** (64-bit). Ensure `gcc` works in the terminal.
* **Error: Windows Defender blocks the file**
* **Cause:** Unsigned Go binaries are often treated as "Unknown" threats.
* **Fix:** Add the project folder to Windows Defender exclusions.
* **Tests fail on `TestButtonPress` (or similar GUI tests)**
* **Cause:** Sometimes Fyne requires a dummy driver for headless testing.
* **Fix:** Ensure the test file sets `test.NewApp()` rather than `app.New()`.

---

# Appendix: The Build Script (`build_and_test.ps1`)

*If the script file is lost, copy this code and save it as `build_and_test.ps1` in the project root.*

```powershell
Write-Host "🧪 1. Running Automated Tests..." -ForegroundColor Cyan

# Run tests recursively. 
# -v gives verbose output.
go test -v ./...

# Check if tests failed ($LASTEXITCODE is non-zero on failure)
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Tests Failed! Build cancelled." -ForegroundColor Red
    exit 1
}

Write-Host "✅ Tests Passed!" -ForegroundColor Green
Write-Host "🔨 2. Building Application..." -ForegroundColor Cyan

# Build the release version (hides console window)
fyne package -release

if ($?) {
    Write-Host "🎉 Success! Application built." -ForegroundColor Green
} else {
    Write-Host "❌ Build Failed." -ForegroundColor Red
}

# Pause to let you read the output before window closes
Read-Host -Prompt "Press Enter to exit"

```
