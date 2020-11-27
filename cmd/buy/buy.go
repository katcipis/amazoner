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
	)

	flag.StringVar(&link, "link", "", "link of product to buy")
	flag.UintVar(&maxPrice, "max", 1000, "max price of product")

	flag.Parse()

	if link == "" {
		fmt.Println("link is an obligatory parameter")
		os.Exit(1)
		return
	}

	fmt.Printf("buy product from link %q max price %d\n\n", link, maxPrice)

	purchase, err := buy.Do(link, maxPrice)
	fmt.Println("==== BUY START ====\n")
	fmt.Println(purchase)
	fmt.Println("==== BUY END ====\n")

	if err != nil {
		logerr("==== ERRORS START ====\n")
		logerr(err.Error())
		logerr("==== ERRORS END ====\n")
	}
}

func logerr(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}
