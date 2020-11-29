package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/katcipis/amazoner/buy"
)

func main() {
	var (
		link        string
		maxPrice    uint
		email       string
		password    string
		userDataDir string
		dryRun      bool
	)

	flag.StringVar(&link, "link", "", "link of product to buy")
	flag.UintVar(&maxPrice, "max", 1000, "max price of product")
	flag.StringVar(&email, "email", "", "your Amazon user email")
	flag.StringVar(&password, "password", "", "your Amazon user password")
	flag.StringVar(&userDataDir, "user-data-dir", "", "your chrome user data dir")
	flag.BoolVar(&dryRun, "dryrun", false, "if true it just opens page without buying")

	flag.Parse()

	if link == "" {
		fmt.Println("link is an obligatory parameter")
		os.Exit(1)
		return
	}

	if userDataDir == "" {
		if email == "" || password == "" {
			fmt.Println("if you are not using user-data-dir, please provider email and password")
			os.Exit(1)
			return
		}
	}

	fmt.Printf("buy product from link %q max price %d\n\n", link, maxPrice)

	fmt.Println("==== BUY START ====")

	purchase, err := buy.Do(link, maxPrice, email, password, userDataDir, dryRun)
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
