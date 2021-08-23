#!/bin/bash

# osx 64bit
# linux 32bit 64bit arm arm64
# windows 32bit 64bit

# TODO ARM support. figure out mapping to uname -m output.

mkdir -p builds
GOOS=darwin GOARCH=amd64 go build -o builds/darwin-x86_64
GOOS=linux GOARCH=386 go build -o builds/linux-i386
GOOS=linux GOARCH=amd64 go build -o builds/linux-x86_64
GOOS=windows GOARCH=386 go build -o builds/windows-i386
GOOS=windows GOARCH=amd64 go build -o builds/windows-x86_64
