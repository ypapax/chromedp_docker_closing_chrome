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

rund(){
  docker build  -t chromedp-alpine .
  docker container run -it --rm --security-opt seccomp=$(pwd)/chrome.json chromedp-alpine
}

"$@"