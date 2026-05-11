package main

import (
	"fmt"
	"log"
	"pandaria/driver"
	"pandaria/httpproxy"
	"sync"

	"github.com/chromedp/chromedp"
)

func start(d driver.Driver, wg *sync.WaitGroup) {
	defer wg.Done()

	err := d.Run(
		chromedp.Navigate("https://www.airbnb.com"),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	//err := godotenv.Load()
	//if err != nil {
	//	log.Fatal("Error loading .env file")
	//}

	// Start HTTP proxy.
	go httpproxy.Run("127.0.0.1:3000")

	// Start Chrome, specifying proxy.
	chrome := driver.InitChrome("127.0.0.1:3000")
	defer chrome.Close()

	// Start Chrome, specifying proxy.
	lp := driver.InitLightpanda("127.0.0.1:3000")
	defer lp.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go start(chrome, &wg)
	go start(lp, &wg)
	wg.Wait()

	fmt.Println("fetch completed")
}
