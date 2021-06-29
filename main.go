package main

import (
	"context"
	"github.com/chromedp/chromedp"
	"log"
	"sync"
	"time"
)

var (
	commonContext *context.Context
	commonContextMtx sync.Mutex
)


func main() {
	log.Println("app is running")
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