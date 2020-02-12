#!/usr/bin/env bash

apps=( narcotk-cf-ddns )

set -x

for i in "${apps[@]}"
do
  GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o binaries/x86_64/$i $i.go; upx binaries/x86_64/$i
  GOOS=linux GOARCH=arm GOARM=6 go build -ldflags "-s -w" -o binaries/rpi/$i $i.go; upx binaries/rpi/$i
  GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o binaries/osx/$i $i.go; upx binaries/osx/$i
  GOOS=linux GOARCH=mipsle go build -ldflags "-s -w" -o binaries/er-mediatek/$i $i.go; upx binaries/er-mediatek/$i
  GOOS=linux GOARCH=mips go build -ldflags "-s -w" -o binaries/er-cavium/$i $i.go; upx binaries/er-cavium/$i
done
