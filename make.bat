@echo off
setlocal enabledelayedexpansion

echo ===================================
echo MCP-DevTools Docker Make Utility
echo ===================================
echo.
echo This utility runs Makefile targets using Docker.
echo.

REM Check if Docker is installed and running
echo Checking Docker status...
docker --version > nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Error: Docker is not installed or not in PATH.
    echo Please install Docker Desktop and try again.
    exit /b 1
)

REM Try a simple Docker command to check if Docker is running
docker info > nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Error: Cannot connect to Docker.
    echo.
    echo Possible solutions:
    echo 1. Make sure Docker Desktop is running
    echo 2. Check if Docker service is started
    echo 3. Restart Docker Desktop
    echo 4. Ensure Docker Desktop is properly installed
    echo.
    echo For more information, see README.docker.md
    exit /b 1
)

echo Docker is running.

REM Check if the Docker image exists, build if not
docker image inspect mcp-devtools-dev > nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Docker image 'mcp-devtools-dev' not found.
    echo Building Docker image...
    docker build -f Dockerfile.dev -t mcp-devtools-dev .
    if %ERRORLEVEL% NEQ 0 (
        echo Error: Failed to build Docker image.
        exit /b 1
    )
    echo Docker image built successfully.
    echo.
)

:menu
echo Available targets:
echo  1. build            - Build for Linux (default)
echo  2. build-windows    - Build for Windows
echo  3. build-macos      - Build for macOS
echo  4. build-all        - Build for all platforms
echo  5. test             - Run tests
echo  6. clean            - Clean build artifacts
echo  7. run-http         - Run server with HTTP transport
echo  8. help             - Show all available targets
echo  9. custom           - Enter a custom target
echo  0. exit             - Exit this utility
echo.

set /p choice=Enter your choice (0-9): 

if "%choice%"=="0" goto :eof
if "%choice%"=="1" set target=build
if "%choice%"=="2" set target=build-windows
if "%choice%"=="3" set target=build-macos
if "%choice%"=="4" set target=build-all
if "%choice%"=="5" set target=test
if "%choice%"=="6" set target=clean
if "%choice%"=="7" set target=run-http
if "%choice%"=="8" set target=help
if "%choice%"=="9" goto custom

if not defined target (
    echo Invalid choice. Please try again.
    echo.
    goto menu
)

goto execute

:custom
echo.
set /p target=Enter custom Makefile target: 
if "!target!"=="" (
    echo No target specified. Please try again.
    echo.
    goto menu
)

:execute
echo.
echo Executing: make %target%
echo.

if "%target%"=="run-http" (
    echo Running with port 18080 exposed...
    docker run --rm -v %CD%:/app -p 18080:18080 mcp-devtools-dev %target%
) else (
    docker run --rm -v %CD%:/app mcp-devtools-dev %target%
)

set exitcode=%ERRORLEVEL%
echo.
if %exitcode% EQU 0 (
    echo Command completed successfully.
) else (
    echo Command failed with exit code %exitcode%.
)

echo.
set target=
set /p continue=Press Enter to continue or type 'exit' to quit: 
if "%continue%"=="exit" goto :eof
echo.
goto menu
