package main

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	log.Println("app is running")
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
		log.Fatal(err)
	}
	log.Printf("result:\n%s", result)
}
