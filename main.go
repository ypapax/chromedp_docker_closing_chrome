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
			ctx0, cancel2 := chromedp.NewContext(
				context.Background(),
				chromedp.WithLogf(log.Printf),
			)
			defer cancel2()
			defer ctx0.Done()

			u := `https://github.com/`
			selector := `title`
			log.Println("requesting", u)
			log.Println("selector", selector)
			var result string
			err := chromedp.Run(ctx0,
				chromedp.Navigate(u),
				chromedp.WaitReady(selector),
				chromedp.OuterHTML(selector, &result),
			)
			if err != nil {
				log.Printf("error %+v \n", err)
			}
			log.Printf("result:\n%s", result)
			if errCancel := chromedp.Cancel(ctx0); errCancel != nil {
				log.Printf("cancel context error: %+v \n", errCancel)
			} else {
				log.Printf("cancel run without an error!")
			}
		}()
		sl := time.Second
		log.Printf("sleeping for %s\n", sl)
		time.Sleep(sl)
	}
}