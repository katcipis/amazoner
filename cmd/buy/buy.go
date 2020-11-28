package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fedesog/webdriver"
	"github.com/katcipis/amazoner/buy"
)

const (
	domain        = "www.amazon.com"
	entrypointURL = "https://" + domain
)

func main() {
	var (
		link     string
		maxPrice uint
		email    string
		password string
	)

	flag.StringVar(&link, "link", "", "link of product to buy")
	flag.UintVar(&maxPrice, "max", 1000, "max price of product")
	flag.StringVar(&email, "email", "", "your Amazon user email")
	flag.StringVar(&password, "password", "", "your Amazon user password")

	flag.Parse()

	if link == "" {
		fmt.Println("link is an obligatory parameter")
		os.Exit(1)
		return
	}

	fmt.Printf("buy product from link %q max price %d\n\n", link, maxPrice)

	fmt.Println("==== BUY START ====\n")

	// Start Chromedriver
	chromeDriver := webdriver.NewChromeDriver("chromedriver")
	err := chromeDriver.Start()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Logging in...")
	session, err := buy.Login(chromeDriver, entrypointURL, email, password)
	if err != nil {
		fmt.Println(err)
	}

	purchase, err := buy.Do(session, link, maxPrice)
	fmt.Println(purchase)
	fmt.Println("==== BUY END ====\n")

	session.Delete()
	chromeDriver.Stop()

	if err != nil {
		logerr("==== ERRORS START ====\n")
		logerr(err.Error())
		logerr("==== ERRORS END ====\n")
	}
}

func logerr(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}
