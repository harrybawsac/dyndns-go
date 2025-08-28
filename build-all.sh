#!/bin/bash
set -e

APP_NAME="dyndns-go"
PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "linux/arm"
  "linux/386"
  "windows/amd64"
  "windows/386"
  "darwin/amd64"
  "darwin/arm64"
  "freebsd/amd64"
  "freebsd/386"
)

mkdir -p bin

for PLATFORM in "${PLATFORMS[@]}"
do
  IFS="/" read -r GOOS GOARCH <<< "$PLATFORM"
  OUTPUT="bin/${APP_NAME}-${GOOS}-${GOARCH}"
  if [ "$GOOS" == "windows" ]; then OUTPUT+='.exe'; fi
  env GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT" main.go
  echo "Built $OUTPUT"
done
