#!/bin/bash
set -e
set -x

echo "Starting Sample JSON rules..."
sample-json --geolite2-db=/opt/Geolite2-City.mmdb
