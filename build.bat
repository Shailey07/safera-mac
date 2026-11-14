@echo off
REM Paste-SEB Injector — build script

echo ============================================
echo   Building Paste-SEB Injector
echo ============================================

REM Dependencies fetch karo
echo [1/3] Fetching dependencies...
go mod tidy
if %errorlevel% neq 0 (
    echo ERROR: go mod tidy failed
    pause
    exit /b 1
)

REM Build (GUI, no console window)
echo [2/3] Building paste-seb.exe (GUI)...
set CGO_ENABLED=1
go build -ldflags="-H windowsgui -s -w" -o paste-seb.exe .
if %errorlevel% neq 0 (
    echo ERROR: build failed
    pause
    exit /b 1
)

REM Debug build (console ke saath, testing ke liye)
echo [3/3] Building paste-seb-debug.exe (console)...
go build -o paste-seb-debug.exe .
if %errorlevel% neq 0 (
    echo WARN: debug build failed (optional)
)

echo.
echo ============================================
echo   Build complete!
echo   - paste-seb.exe        (GUI version)
echo   - paste-seb-debug.exe  (console version)
echo ============================================
pause