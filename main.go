package main

import (
	"context"
	"github.com/chromedp/chromedp"
	"log"
	"time"
)

func main() {
	log.Println("app is running")
	cycle := 0
	for {
		cycle++
		log.Printf("starting cycle %+v\n", cycle)
		func(){
			log.SetFlags(log.LstdFlags | log.Llongfile)
			ctx, cancel := chromedp.NewContext(
				context.Background(),
				chromedp.WithLogf(log.Printf),
			)
			defer cancel()

			// create a timeout
			ctx, cancel = context.WithTimeout(ctx, 15 * time.Second)
			defer cancel()

			u := `https://www.whatismybrowser.com/detect/what-is-my-user-agent`
			selector := `#detected_value`
			log.Println("requesting", u)
			log.Println("selector", selector)
			var result string
			err := chromedp.Run(ctx,
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
