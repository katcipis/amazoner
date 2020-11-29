package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/katcipis/amazoner/buy"
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

	fmt.Println("==== BUY START ====")

	purchase, err := buy.Do(link, maxPrice, email, password)
	fmt.Println(purchase)
	fmt.Println("==== BUY END ====")

	if err != nil {
		logerr("==== ERRORS START ====")
		logerr(err.Error())
		logerr("==== ERRORS END ====")
	}
}

func logerr(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}
