#!/usr/bin/env bash

PLATFORMS=(
	"darwin:amd64"
	"linux:386"
	"linux:amd64"
	"linux:arm"
	"linux:arm64"
)

for PLAT in "${PLATFORMS[@]}"
do
	GOOS=$(cut -d ":" -f 1 <<< $PLAT)
	GOARCH=$(cut -d ":" -f 2 <<< $PLAT)
	OUTPUT="godot-${GOOS}-${GOARCH}"
	echo BUILDING: $OUTPUT
	env GOOS=$GOOS GOARCH=$GOARCH go build -o "build/$OUTPUT"
done
