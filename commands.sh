#!/usr/bin/env bash
set -ex
set -o pipefail

run(){
  docker-compose build
  docker-compose up
}

loca(){
  go run main.go
}

runChrome(){
#  /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --kiosk
  /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome -remote-debugging-port 9222
#  chrome
}

check(){
  container=$(docker ps | grep chrome | awk '{  print $10 }')
  docker exec -ti $container /bin/bash
  # and then paste: ps -aux | grep chrome | wc -l
}

"$@"