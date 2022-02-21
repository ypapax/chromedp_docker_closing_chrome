package main

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/ypapax/logrus_conf"
	"log"
	"sync"
	"time"
)

var (
	commonContext *context.Context
	commonContextMtx sync.Mutex
)

func main(){
	mainCommonContext()
}

func mainWithCancelContext() {
	if err := func() error {
		if err := logrus_conf.PrepareFromEnv("chromedp_docker_closing_chrome"); err != nil {
			return errors.WithStack(err)
		}
		log.Println("app is running")
		cycle := 0
		for {
			cycle++
			log.Printf("starting cycle %+v\n", cycle)
			func(){
				log.SetFlags(log.LstdFlags | log.Llongfile)
				ctx0, _ := chromedp.NewContext(
					context.Background(),
					chromedp.WithLogf(log.Printf),
				)
				//defer cancel()


				u := `https://github.com/`
				selector := `title`
				log.Println("requesting", u)
				log.Println("selector", selector)
				var result string
				err := chromedp.Run(ctx0,
					chromedp.Navigate(u),
					chromedp.WaitReady(selector),
					chromedp.OuterHTML(selector, &result),
					//chromedp.Stop()
				)
				if err != nil {
					log.Printf("error %+v \n", err)
				}
				log.Printf("result:\n%s", result)
			}()
			sl := 1 * time.Second
			log.Printf("sleeping for %s\n", sl)
			if cycle > 20 {
				logrus.Infof("enough")
				return nil
			}
			time.Sleep(sl)
		}
	}(); err != nil {
		logrus.Errorf("%+v", err)
	}
	logrus.Infof("sleeping forever")
	select {}
}



/*func killZombies() error {

}

func zombieProcessIds(inpContext context.Context) ([]int, error) {
	ctx, cancel := context.WithTimeout(inpContext, 5 * time.Minute)
	defer cancel()
	zombieMark := map[string]bool {
		"Z": true,
		"Z+": true,
	}
	//"ps  xao pid,stat"
	cmd := exec.CommandContext(ctx, "ps", "xao", "pid,stat")
	b, err := cmd.Output()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(b) == 0 {
		logrus.Warnf("no output")
		return nil, nil
	}
	lines := strings.Split(string(b), "\n")
	var zombiePids []int
	for _, l := range lines {
		leftRight := strings.Split(l, " ")
		if len(leftRight) != 2 {
			return nil, errors.Errorf("not correct parts")
		}
		right := leftRight[1]
		if _, ok := zombieMark[right]; !ok {
			continue
		}
		left := leftRight[0]
		pid, errA := strconv.Atoi(left)
		if errA != nil {
			return nil, errors.WithStack(errA)
		}
		zombiePids
	}
}
*/

func mainCommonContext() {
	log.Println("app is running withCommon context")
	cycle := 0
	for {
		cycle++
		log.Printf("starting cycle %+v\n", cycle)
		func(){
			commonContextMtx.Lock()
			defer commonContextMtx.Unlock()
			log.SetFlags(log.LstdFlags | log.Llongfile)
			if commonContext == nil {
				ctx0, _ := chromedp.NewContext(
					context.Background(),
					chromedp.WithLogf(log.Printf),
				)
				commonContext = &ctx0
			}


			u := `https://github.com/`
			selector := `title`
			log.Println("requesting", u)
			log.Println("selector", selector)
			var result string
			err := chromedp.Run(*commonContext,
				chromedp.Navigate(u),
				chromedp.WaitReady(selector),
				chromedp.OuterHTML(selector, &result),
			)
			if err != nil {
				log.Printf("error %+v \n", err)
			}
			log.Printf("result:\n%s", result)
		}()
		sl := time.Second
		log.Printf("sleeping for %s\n", sl)
		time.Sleep(sl)
	}
}