#!/usr/bin/env bash
set -ex
set -o pipefail

run(){
  docker-compose build
  docker-compose up
}

runl(){
  go run main.go
}

buildLinux(){
  bin=~/tmp/chromedb_linux
  GOOS=linux GOARCH=amd64  go build -o $bin
  ls -laht $bin
}

rund(){
  docker build  -t chromedp-alpine .
  docker container run -it --rm --security-opt seccomp=$(pwd)/chrome.json chromedp-alpine
}

"$@"