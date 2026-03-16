@echo off
setlocal EnableExtensions EnableDelayedExpansion
echo.
echo ===========================================
echo  Data Coupler - Build Script
echo ===========================================
echo.

:: ── Version ────────────────────────────────────────────────────────────────
SET VERSION=0.4.0

:: Keep Go temp/cache inside the repo so builds do not depend on AppData permissions.
set "GOTMPDIR=%CD%\.tmp\go-build"
set "GOCACHE=%CD%\.tmp\gocache"
set "GOMODCACHE=%CD%\.tmp\gomodcache"
set "TEMP=%CD%\.tmp\temp"
set "TMP=%CD%\.tmp\temp"
set "TMPDIR=%CD%\.tmp\temp"
if not exist "%CD%\.tmp" mkdir "%CD%\.tmp"
if not exist "%GOTMPDIR%" mkdir "%GOTMPDIR%"
if not exist "%GOCACHE%" mkdir "%GOCACHE%"
if not exist "%GOMODCACHE%" mkdir "%GOMODCACHE%"
if not exist "%TEMP%" mkdir "%TEMP%"

:: Prefer MSYS2 UCRT64 for cgo/Fyne builds instead of any older GCC on PATH.
set "MSYS2_ROOT="
if exist "C:\Dev\msys64\ucrt64\bin\gcc.exe" (
    set "MSYS2_ROOT=C:\Dev\msys64"
) else (
    if exist "C:\msys64\ucrt64\bin\gcc.exe" (
        set "MSYS2_ROOT=C:\msys64"
    )
)

if defined MSYS2_ROOT (
    set "CC=!MSYS2_ROOT!\ucrt64\bin\gcc.exe"
    set "CXX=!MSYS2_ROOT!\ucrt64\bin\g++.exe"
    set "PATH=!MSYS2_ROOT!\ucrt64\bin;!MSYS2_ROOT!\usr\bin;!PATH!"
    echo Using MSYS2 compiler: !CC!
    echo.
)

if not exist "Icon.png" (
    echo Generating placeholder Icon.png...
    powershell -NoProfile -Command ^
        "Add-Type -AssemblyName System.Drawing;" ^
        "$b=New-Object System.Drawing.Bitmap 256,256;" ^
        "$g=[System.Drawing.Graphics]::FromImage($b);" ^
        "$g.FillRectangle([System.Drawing.Brushes]::SteelBlue,0,0,256,256);" ^
        "$g.Dispose();$b.Save('Icon.png');$b.Dispose()"
    echo Icon.png created.
    echo.
)

echo [1/4] Running tests...
go test ./...
if %ERRORLEVEL% neq 0 (
    echo.
    echo FAILED: Tests did not pass. Build cancelled.
    pause
    exit /b 1
)
echo Tests passed.
echo.

echo [2/4] Building GUI executable (version %VERSION%)...
:: TDM-GCC's older binutils can emit broken DWARF sections with Go 1.25+ cgo builds on Windows.
:: Stripping debug info avoids the invalid executable layout and produces a working release binary.
go build -ldflags="-s -w -H=windowsgui -X github.com/mitchellbauer/data-coupler/internal/version.AppVersion=%VERSION%" -o "Data Coupler.exe" ./cmd/datacoupler
if %ERRORLEVEL% neq 0 (
    echo.
    echo FAILED: Build failed.
    pause
    exit /b 1
)
echo Binary built: Data Coupler.exe
echo Output written to: %CD%\Data Coupler.exe
echo.

echo [3/4] Code signing...
IF EXIST certificate.pfx (
    echo Signing executable...
    signtool sign /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 /f certificate.pfx /p %SIGN_PASSWORD% "Data Coupler.exe"
    if %ERRORLEVEL% neq 0 (
        echo WARNING: Signing failed. Binary was built but is unsigned.
    ) else (
        echo Signed successfully.
    )
) ELSE (
    echo Skipping signing (certificate.pfx not found).
)
echo.

echo [4/4] MSI installer...
where wix >nul 2>&1
IF %ERRORLEVEL%==0 (
    if not exist "dist" mkdir "dist"
    echo Building MSI installer...
    wix build installer\windows\DataCoupler.wxs -d Version=%VERSION% -o "dist\DataCoupler-%VERSION%.msi"
    if %ERRORLEVEL% neq 0 (
        echo WARNING: MSI build failed.
    ) else (
        echo MSI built: dist\DataCoupler-%VERSION%.msi
        IF EXIST certificate.pfx (
            signtool sign /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 /f certificate.pfx /p %SIGN_PASSWORD% "dist\DataCoupler-%VERSION%.msi"
        )
    )
) ELSE (
    echo Skipping MSI (wix tool not found).
    echo To enable: dotnet tool install --global wix
)
echo.

echo Done.
echo.
pause
