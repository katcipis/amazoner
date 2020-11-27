package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/katcipis/amazoner/search"
)

func main() {
	var (
		name     string
		minPrice uint
		maxPrice uint
	)

	flag.StringVar(&name, "name", "", "name of product")
	flag.UintVar(&minPrice, "min", 0, "min price of product")
	flag.UintVar(&maxPrice, "max", 10000, "max price of product")

	flag.Parse()

	if name == "" {
		fmt.Println("name is an obligatory parameter")
		os.Exit(1)
		return
	}

	fmt.Printf("search product %q min price %d max price %d\n\n", name, minPrice, maxPrice)

	for _, res := range search.Do(name, minPrice, maxPrice) {
		fmt.Println(res)
	}
}
