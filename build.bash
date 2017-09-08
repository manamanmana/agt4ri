#!/bin/bash
set -eu

OUTDIR=bin

GOOS=linux GOARCH=amd64 go build -o "${OUTDIR}/linux/agt4ri"
GOOS=darwin GOARCH=amd64 go build -o "${OUTDIR}/osx/agt4ri"
GOOS=windows GOARCH=amd64 go build -o "${OUTDIR}/win/agt4ri.exe"

exit 0

