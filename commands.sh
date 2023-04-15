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

check(){
  container=$(docker ps | grep chrome | awk '{  print $10 }')
  docker exec -ti $container /bin/bash
  # and then paste:
  #                 ps -aux | grep chrome | wc -l

}

checkLoca(){
  ps  | grep chrome 
  ps  | grep chrome | wc -l
}

"$@"