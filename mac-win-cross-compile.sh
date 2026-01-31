#! /bin/bash
# Copyright (c) Tim Lyakhovetskiy
# SPDX-License-Identifier: MPL-2.0

set -ex

# env GOOS=windows GOARCH=amd64 go clean -modcache

# This command expects mingw32 to be installed, openal-soft to be installed,
# and the Windows version of OpenAL-soft (specifically the router DLL) to be in
# /opt/local/lib/OpenAL32.dll
export GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX='x86_64-w64-mingw32-g++' CC='x86_64-w64-mingw32-gcc' CGO_CFLAGS='-I/opt/local/include' CGO_LDFLAGS='-L/opt/local/lib'
go build -o ./bin/game-win.exe ./game/
go build -o ./bin/editor-win.exe ./editor/
# cp ./bin/*.exe /Volumes/Archive/Temp/gofoom
# cp -r ./data/worlds /Volumes/Archive/Temp/gofoom/data
scp ./bin/*.exe root@jane.home:/mnt/tank/Archive/Temp/gofoom/

