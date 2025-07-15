# Docker Development Environment for mcp-devtools

This document explains how to use the Docker development environment for building and running mcp-devtools, especially useful for Windows users who want to use Makefile functionality without installing Unix tools directly.

## Quick Start for Windows Users

A batch file utility `make.bat` is provided for Windows users to easily run Makefile targets:

1. Simply run `make.bat` from the command prompt:
   ```
   make.bat
   ```

2. Select an option from the menu or enter a custom Makefile target.

This utility will:
- Check if Docker is installed
- Build the Docker image if needed
- Run the selected Makefile target inside Docker
- Handle volume mounting automatically

## Prerequisites

- Docker installed on your system
- Git (for cloning the repository)

## Setup

1. Clone the repository:
   ```
   git clone <repository-url>
   cd mcp-devtools
   ```

2. Build the development Docker image:
   ```
   docker build -f Dockerfile.dev -t mcp-devtools-dev .
   ```

## Usage

### Building for Windows

To build a Windows executable (.exe):

```
docker run -v ${PWD}:/app mcp-devtools-dev build-windows
```

The Windows executable will be created in the `bin/` directory as `mcp-devtools.exe`.

### Building for All Platforms

To build executables for Linux, Windows, and macOS:

```
docker run -v ${PWD}:/app mcp-devtools-dev build-all
```

This will create:
- `bin/mcp-devtools` (Linux)
- `bin/mcp-devtools.exe` (Windows)
- `bin/mcp-devtools-macos` (macOS)

### Running Other Make Targets

You can run any Make target available in the Makefile:

```
docker run -v ${PWD}:/app mcp-devtools-dev <target>
```

For example:
- `docker run -v ${PWD}:/app mcp-devtools-dev test` - Run tests
- `docker run -v ${PWD}:/app mcp-devtools-dev clean` - Clean build artifacts
- `docker run -v ${PWD}:/app mcp-devtools-dev help` - Show available targets

### Running the Server

To run the server with HTTP transport:

```
docker run -v ${PWD}:/app -p 18080:18080 mcp-devtools-dev run-http
```

## Windows Command Prompt Notes

If you're not using the `make.bat` utility and prefer to run Docker commands directly, use `%CD%` instead of `${PWD}` in Command Prompt:

```
docker run -v %CD%:/app mcp-devtools-dev build-windows
```

## Windows PowerShell Notes

If you're not using the `make.bat` utility and prefer to run Docker commands directly in PowerShell, you may need to use the full path:

```
docker run -v ${PWD}:/app mcp-devtools-dev build-windows
```

Or with explicit path conversion:

```
docker run -v ${PWD -replace '\\', '/'}:/app mcp-devtools-dev build-windows
```

## Troubleshooting

### Docker Connection Issues

If you see errors like:
```
ERROR: error during connect: Head "http://%2F%2F.%2Fpipe%2FdockerDesktopLinuxEngine/_ping": open //./pipe/dockerDesktopLinuxEngine: The system cannot find the file specified.
```

This indicates Docker Desktop is not running or not properly configured. Try these solutions:

1. **Start Docker Desktop**: Make sure Docker Desktop application is running
2. **Check Docker Service**: Ensure the Docker service is started in Windows Services
3. **Restart Docker**: Right-click the Docker icon in the system tray and select "Restart"
4. **Reinstall Docker**: If problems persist, consider reinstalling Docker Desktop

### Permission Issues

If you encounter permission issues with the mounted volume, you may need to adjust the permissions or use a different volume mounting strategy.

### Path Issues

Windows paths with spaces may cause issues. Try to use a repository path without spaces, or use the appropriate quoting for your shell.
