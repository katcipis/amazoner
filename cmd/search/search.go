package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/katcipis/amazoner/product"
	"github.com/katcipis/amazoner/search"
)

func main() {
	var (
		domain   string
		name     string
		minPrice uint
		maxPrice uint
		filter   bool
	)

	flag.StringVar(&domain, "domain", "www.amazon.com", "Amazon domain to search")
	flag.StringVar(&name, "name", "", "name of product")
	flag.UintVar(&minPrice, "min", 0, "min price of product")
	flag.UintVar(&maxPrice, "max", 10000, "max price of product")
	flag.BoolVar(&filter, "filter", false, "filter results")

	flag.Parse()

	if name == "" {
		fmt.Println("name is an obligatory parameter")
		os.Exit(1)
		return
	}

	fmt.Printf("search product %q min price %d max price %d\n\n", name, minPrice, maxPrice)

	searcher := search.New(time.Second)
	products, err := searcher.Search(domain, name, minPrice, maxPrice)

	if filter {
		products = product.Filter(name, products)
	}

	fmt.Println("==== RESULTS START ====")
	for _, prod := range products {
		fmt.Printf("%+v\n", prod)
	}
	fmt.Println("==== RESULTS END ====")

	if err != nil {
		logerr("==== ERRORS START ====")
		logerr(err.Error())
		logerr("==== ERRORS END ====")
	}
}

func logerr(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}
