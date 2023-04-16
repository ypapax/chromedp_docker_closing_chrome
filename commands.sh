#!/usr/bin/env bash
set -ex
set -o pipefail

run(){
  docker-compose build
  docker-compose up
}

loca(){
  out=/tmp/chromedp_docker_closing_chrome
  go build -o $out
  $out
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
# number of opened files:
#lsof -Fn | sort | uniq | wc -l


"$@"