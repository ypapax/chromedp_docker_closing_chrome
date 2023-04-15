package main

import (
	"github.com/pkg/errors"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"log"
	"time"
)

func main() {
	log.Println("app is running")
	cycle := 0
	LogChromeMem()
	for {
		cycle++
		log.Printf("starting cycle %+v\n", cycle)
		if err := func() (finalErr error) {
			log.SetFlags(log.LstdFlags | log.Llongfile)

			// Run Chrome browser
			service, err := selenium.NewChromeDriverService("./chromedriver", 4444)
			if err != nil {
				return errors.WithStack(err)
			}
			defer func(){
				if errS := service.Stop(); errS != nil {
					if finalErr != nil {
						finalErr = errors.WithStack(errS)
					} else {
						log.Printf("error: %+v", errS)
					}

				}
			}()

			caps := selenium.Capabilities{}
			caps.AddChrome(chrome.Capabilities{Args: []string{
				"window-size=1920x1080",
				"--no-sandbox",
				"--disable-dev-shm-usage",
				"disable-gpu",
				"--headless",  // comment out this line to see the browser
			}})

			driver, err := selenium.NewRemote(caps, "")
			if err != nil {
				return errors.WithStack(err)
			}

			if errG := driver.Get("https://github.com"); errG != nil {
				return errors.WithStack(errG)
			}
			title, err := driver.Title()
			if err != nil {
				return errors.WithStack(err)
			}
			log.Printf("title: %+v", title)
			if errC := driver.Close(); errC != nil {
				return errors.WithStack(errC)
			}
			if errQ := driver.Quit(); errQ != nil {
				return errors.WithStack(errQ)
			}
			return nil
		}(); err != nil {
			log.Printf("error: %+v", err)
		} else {
			log.Printf("no errors")
		}
		sl := time.Second
		log.Printf("sleeping for %s\n", sl)
		time.Sleep(sl)
	}
}