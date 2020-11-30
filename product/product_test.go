package product_test

import (
	"testing"

	"github.com/katcipis/amazoner/product"
)

func TestProductGet(t *testing.T) {
	urls := []string{
		"https://www.amazon.com/MSI-GeForce-RTX-2060-Architecture/dp/B07MQ36Z6L",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			p, err := product.Get(url)
			if err != nil {
				t.Fatal(err)
			}
			if p.Name == "" {
				t.Error("missing name on product")
			}
			if p.Price <= 0 {
				t.Error("missing price on product")
			}
		})
	}
}
