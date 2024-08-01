#! /bin/bash

# env GOOS=windows GOARCH=amd64 go clean -modcache
env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build -o ./bin/game-win.exe ./game/
env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build -o ./bin/editor-win.exe ./editor/
cp ./bin/*.exe /Volumes/Archive/Temp/