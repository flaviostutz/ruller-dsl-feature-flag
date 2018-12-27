#!/bin/bash
set -e
set -x

echo "Starting Sample Rules..."
sample-rules \
    --log-level=$LOG_LEVEL
