package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"log"
	"sync"
	"time"
)

var (
	usedPorts     = make(map[int]struct{})
	usedPortsMtx  = sync.Mutex{}
	freePortsFrom = 4444
	freePortsTo   = 5555
)

func chooseAport() (choosedPort int) {
	defer func(){
		log.Printf("port is choosed: %+v", choosedPort)
	}()
	usedPortsMtx.Lock()
	defer usedPortsMtx.Unlock()
	for {
		for i := freePortsFrom; i <= freePortsTo; i++ {
			_, used := usedPorts[i]
			if used {
				continue
			}
			usedPorts[i] = struct{}{}
			return i
		}
		time.Sleep(time.Second)
	}
}

func freeUpAport(port int) {
	usedPortsMtx.Lock()
	defer usedPortsMtx.Unlock()
	delete(usedPorts, port)
	log.Printf("port is freed up: %+v", port)
}

func seleniumRunChrome(u string) (finalErr error) {
	port := chooseAport()
	service, err := selenium.NewChromeDriverService("./chromedriver", port)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if errS := service.Stop(); errS != nil {
			if finalErr != nil {
				finalErr = errors.WithStack(errS)
			} else {
				log.Printf("error: %+v", errS)
			}
		}
		freeUpAport(port)
	}()

	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{Args: []string{
		"window-size=1920x1080",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"disable-gpu",
		"--headless", // comment out this line to see the browser
	}})

	driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://127.0.0.1:%+v/wd/hub", port))
	if err != nil {
		return errors.WithStack(err)
	}

	if errG := driver.Get(u); errG != nil {
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
}

func seleniumRunFirefox(u string) (finalErr error) {
	port := chooseAport()
	service, err := selenium.NewChromeDriverService("./chromedriver", port)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if errS := service.Stop(); errS != nil {
			if finalErr != nil {
				finalErr = errors.WithStack(errS)
			} else {
				log.Printf("error: %+v", errS)
			}
		}
		freeUpAport(port)
	}()

	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{Args: []string{
		"window-size=1920x1080",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"disable-gpu",
		"--headless", // comment out this line to see the browser
	}})

	driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://127.0.0.1:%+v/wd/hub", port))
	if err != nil {
		return errors.WithStack(err)
	}

	if errG := driver.Get(u); errG != nil {
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
}
