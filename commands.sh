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
#  --security-opt seccomp=$(pwd)/chrome.json
  docker container run -it --rm --name chromedp_docker  chromedp-alpine
}

"$@"