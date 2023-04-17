#!/usr/bin/env bash
set -ex
set -o pipefail

run(){
  docker-compose build
  docker-compose up
}

stats_dp(){
  docker
}

loca(){
  out=/tmp/chromedp_docker_closing_chrome
  go build -o $out
  $out
}

locadp(){
  TYPE=chromedp loca
}

locaselenium(){
  TYPE=selenium loca
}

chromeCont(){
#  contName chromedp_1
  echo chromedp_docker_closing_chrome_chromedp_1
}

seleniumCont(){
#  contName selenium
  echo chromedp_docker_closing_chrome_selenium_1
}

contName(){
  docker ps | grep $1 | awk '{  print $10 }'
}

chromeStats(){
  docker logs $(chromeCont)
}

selStats(){
  docker logs $(seleniumCont)
}

check(){
  container=$(chromeCont)
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