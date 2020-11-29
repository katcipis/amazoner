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
		dryRun   bool
	)

	flag.StringVar(&link, "link", "", "link of product to buy")
	flag.UintVar(&maxPrice, "max", 1000, "max price of product")
	flag.StringVar(&email, "email", "", "your Amazon user email")
	flag.StringVar(&password, "password", "", "your Amazon user password")
	flag.BoolVar(&dryRun, "dryrun", false, "if true it just opens page without buying")

	flag.Parse()

	if link == "" || email == "" || password == "" {
		fmt.Println("link, email and password are obligatory parameters")
		os.Exit(1)
		return
	}

	fmt.Printf("buy product from link %q max price %d\n\n", link, maxPrice)

	fmt.Println("==== BUY START ====")

	purchase, err := buy.Do(link, maxPrice, email, password, dryRun)
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
