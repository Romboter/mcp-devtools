#!/bin/sh
set -e

# If no arguments, show help
if [ $# -eq 0 ]; then
    echo "Usage: docker run mcp-devtools-dev [make-target]"
    echo "Available targets:"
    make -f /app/Makefile help
    exit 1
fi

# Run make with the provided target
cd /app
exec make "$@"
