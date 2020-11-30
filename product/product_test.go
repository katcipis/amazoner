package product_test

import (
	"testing"

	"github.com/katcipis/amazoner/product"
)

func TestProductGet(t *testing.T) {
	// Usually isolated parsing tests would be more reliable and faster
	// But right now corners are being cut :-)
	urls := []string{
		"https://www.amazon.com/MSI-Twin-Frozr-Architecture-Overclocked-Graphics/dp/B07YXPVBWW",
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
