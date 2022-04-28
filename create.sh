#!/usr/bin/env bash
./update
name=${PWD##*/}
go get -u all

GOOS=linux GOARCH="amd64" go build -ldflags="-s -w" -o linux/amd64/"$name"
GOOS=linux GOARCH="arm64" go build -ldflags="-s -w" -o linux/arm64/"$name"

upx --best --lzma linux/amd64/"$name"
upx --best --lzma linux/arm64/"$name"

docker rmi -f petrjahoda/"$name":latest
docker buildx build -t petrjahoda/"$name":latest --platform=linux/arm64,linux/amd64 . --push
docker rmi -f petrjahoda/"$name":2022.2.1
docker buildx build -t petrjahoda/"$name":2022.2.1 --platform=linux/arm64,linux/amd64 . --push
