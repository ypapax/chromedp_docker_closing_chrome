#!/usr/bin/env bash
set -ex
set -o pipefail

now(){
	date '+%Y-%m-%d--%H-%M-%S'
}

run(){
  docker-compose build
  SIMULT=${SIMULT-50} docker-compose up 2>&1 | tee ~/chromedp_sel_$(now).log
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

locaselenium(){
  SIMULT=1\
  TYPE=seleniumff DRIVER_PATH="./geckodriver_mac" loca
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

#selOpenFiles(){
#  check
#}

check(){
  container=$1
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