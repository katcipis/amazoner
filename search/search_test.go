package search_test

import (
	"fmt"
	"testing"

	"github.com/katcipis/amazoner/product"
	"github.com/katcipis/amazoner/search"
)

func TestSearch(t *testing.T) {
	// The best we can do is some sort of property conservation
	// test, so we catch bizarre regressions like returning no results
	// or results with empty name, etc (although we dont check specific
	// products or relevance).

	type Test struct {
		domain     string
		search     string
		minPrice   uint
		maxPrice   uint
		minResults uint
	}

	tests := []Test{
		{
			domain:     "www.amazon.com",
			search:     "nvidia rtx 3070",
			minPrice:   500,
			maxPrice:   1500,
			minResults: 8,
		},
	}

	for _, test := range tests {
		testname := fmt.Sprintf(
			"%s%sMin%dMax%dWant%d",
			test.domain,
			test.search,
			test.minPrice,
			test.maxPrice,
			test.minResults,
		)
		t.Run(testname, func(t *testing.T) {
			res, err := search.Do(test.domain, test.search, test.minPrice, test.maxPrice)
			if len(res) < int(test.minResults) {
				t.Errorf("got %d results; want %d", len(res), test.minResults)
				t.Errorf("results:%v", res)
				if err != nil {
					t.Errorf("errors:%v", err)
				}
			}
		})
	}
}

func TestFilter(t *testing.T) {
	searchResults := []search.Result{
		search.Result{
			Product: product.Product{
				Name:  "MSI GeForce RTX 3070 Ventus 3X OC Gaming Video Card, 8GB GDDR6, PCIe 4.0, 8K, VR Ready, Ray Tracing, 1x HDMI 2.1, 3X DisplayPort 1.4, Triple Fans, HDCP, Mytrix HDMI 2.1 8K Cable",
				Price: 939.99},
			URL: "https://www.amazon.com/MSI-RTX-3070-HDMI-DisplayPort/dp/B08MVFMN35"},
		search.Result{
			Product: product.Product{
				Name:  "MSI GeForce RTX 3070 Ventus 3X OC Gaming Video Card, 8GB GDDR6, PCIe 4.0, 8K, VR Ready, Ray Tracing, 1x HDMI 2.1, 3X DisplayPort 1.4, Triple Fans, HDCP, Battlefield V",
				Price: 939.99},
			URL: "https://www.amazon.com/MSI-RTX-3070-DisplayPort-Battlefield/dp/B08MVHD9Z9"},
		search.Result{
			Product: product.Product{
				Name:  "CyberpowerPC Gamer Xtreme VR Gaming PC, Intel i5-10400F 2.9GHz, GeForce GTX 1660 Super 6GB, 8GB DDR4, 500GB NVMe SSD, WiFi Ready & Win 10 Home (GXiVR8060A10)",
				Price: 799.99},
			URL: "https://www.amazon.com/CyberpowerPC-Xtreme-i5-10400F-GeForce-GXiVR8060A10/dp/B08FBK2DK5"},
		search.Result{
			Product: product.Product{
				Name:  "MSI Gaming GeForce RTX 3070 8GB GDRR6 256-Bit HDMI/DP Tri-Frozr 2 TORX Fan 4.0 Ampere Architecture RGB OC Graphics Card (RTX 3070 Gaming X Trio)",
				Price: 959},
			URL: "https://www.amazon.com/MSI-GeForce-RTX-3070-Architecture/dp/B08KWN2LZG"},
		search.Result{
			Product: product.Product{
				Name:  "MSI Gaming GeForce RTX 3070 8GB GDRR6 256-Bit HDMI/DP TORX Fan 3.0 Ampere Architecture OC Graphics Card (RTX 3070 Ventus 3X OC)",
				Price: 909.99},
			URL: "https://www.amazon.com/MSI-GeForce-256-Bit-Architecture-Graphics/dp/B08KWLMZV4"},
		search.Result{
			Product: product.Product{
				Name:  "PNY GeForce RTX 3070 8GB XLR8 Gaming Revel Epic-X RGB Triple Fan Graphics Card",
				Price: 949.99},
			URL: "https://www.amazon.com/PNY-GeForce-Gaming-Epic-X-Graphics/dp/B08HBJB7YD"},
		search.Result{
			Product: product.Product{
				Name:  "ARESGAME 750W Power Supply Semi Modular 80+ Bronze PSU (AGV750)",
				Price: 79.99},
			URL: "https://www.amazon.com/ARESGAME-Supply-Modular-Bronze-AGV750/dp/B08JM12SQ5"},
		search.Result{
			Product: product.Product{
				Name:  "MSI GeForce RTX 3070 Ventus 3X OC Gaming Video Card, 8GB GDDR6, PCIe 4.0, 8K, VR Ready, Ray Tracing, 1x HDMI 2.1, 3X DisplayPort 1.4, Triple Fans, HDCP, Mytrix HDMI 2.1 8K Cable, Battlefield V",
				Price: 939.99},
			URL: "https://www.amazon.com/MSI-RTX-3070-DisplayPort-Battlefield/dp/B08MVH3QJF"},
		search.Result{
			Product: product.Product{
				Name:  "Beelink U57 Mini PC with Intel Core i5-5257u Processor(up to 3.10 GHz)&Windows 10 Pro,8G DDR3L/256G SSD High Performance Business Mini Computer,2.4G/5G Dual WiFi,BT4.2,Dual HDMI Ports",
				Price: 379},
			URL: "https://www.amazon.com/Beelink-U57-Processor-256G-Performance/dp/B0879KKTCB"},
		search.Result{
			Product: product.Product{
				Name:  "EVGA 08G-P5-3767-KR GeForce RTX 3070 FTW3 Ultra Gaming, 8GB GDDR6, iCX3 Technology, ARGB LED, Metal Backplate",
				Price: 999.99},
			URL: "https://www.amazon.com/EVGA-08G-P5-3767-KR-GeForce-Technology-Backplate/dp/B08L8L9TCZ"},
	}

	type Test struct {
		search         string
		results        []search.Result
		expectedLength int
	}

	tests := []Test{
		{
			search:         "rtx 3070",
			results:        searchResults,
			expectedLength: 7,
		},
	}

	for _, test := range tests {
		testname := fmt.Sprintf(
			"%sLen%d",
			test.search,
			test.expectedLength,
		)
		t.Run(testname, func(t *testing.T) {
			filteredResults := search.Filter(test.search, test.results)
			res := len(filteredResults)

			if res != test.expectedLength {
				t.Errorf("got length %d; want %d", res, test.expectedLength)
				prettyPrintResultList("results", test.results, t)
				prettyPrintResultList("filteredResults", filteredResults, t)
			}
		})
	}
}

func TestSortByPrice(t *testing.T) {
	searchResults := []search.Result{
		search.Result{
			Product: product.Product{
				Name:  "MSI GeForce RTX 3070 Ventus 3X OC Gaming Video Card, 8GB GDDR6, PCIe 4.0, 8K, VR Ready, Ray Tracing, 1x HDMI 2.1, 3X DisplayPort 1.4, Triple Fans, HDCP, Mytrix HDMI 2.1 8K Cable",
				Price: 939.99},
			URL: "https://www.amazon.com/MSI-RTX-3070-HDMI-DisplayPort/dp/B08MVFMN35"},
		search.Result{
			Product: product.Product{
				Name:  "MSI GeForce RTX 3070 Ventus 3X OC Gaming Video Card, 8GB GDDR6, PCIe 4.0, 8K, VR Ready, Ray Tracing, 1x HDMI 2.1, 3X DisplayPort 1.4, Triple Fans, HDCP, Battlefield V",
				Price: 939.99},
			URL: "https://www.amazon.com/MSI-RTX-3070-DisplayPort-Battlefield/dp/B08MVHD9Z9"},
		search.Result{
			Product: product.Product{
				Name:  "MSI Gaming GeForce RTX 3070 8GB GDRR6 256-Bit HDMI/DP Tri-Frozr 2 TORX Fan 4.0 Ampere Architecture RGB OC Graphics Card (RTX 3070 Gaming X Trio)",
				Price: 959},
			URL: "https://www.amazon.com/MSI-GeForce-RTX-3070-Architecture/dp/B08KWN2LZG"},
		search.Result{
			Product: product.Product{
				Name:  "MSI Gaming GeForce RTX 3070 8GB GDRR6 256-Bit HDMI/DP TORX Fan 3.0 Ampere Architecture OC Graphics Card (RTX 3070 Ventus 3X OC)",
				Price: 909.99},
			URL: "https://www.amazon.com/MSI-GeForce-256-Bit-Architecture-Graphics/dp/B08KWLMZV4"},
		search.Result{
			Product: product.Product{
				Name:  "PNY GeForce RTX 3070 8GB XLR8 Gaming Revel Epic-X RGB Triple Fan Graphics Card",
				Price: 949.99},
			URL: "https://www.amazon.com/PNY-GeForce-Gaming-Epic-X-Graphics/dp/B08HBJB7YD"},
		search.Result{
			Product: product.Product{
				Name:  "MSI GeForce RTX 3070 Ventus 3X OC Gaming Video Card, 8GB GDDR6, PCIe 4.0, 8K, VR Ready, Ray Tracing, 1x HDMI 2.1, 3X DisplayPort 1.4, Triple Fans, HDCP, Mytrix HDMI 2.1 8K Cable, Battlefield V",
				Price: 939.99},
			URL: "https://www.amazon.com/MSI-RTX-3070-DisplayPort-Battlefield/dp/B08MVH3QJF"},
		search.Result{
			Product: product.Product{
				Name:  "EVGA 08G-P5-3767-KR GeForce RTX 3070 FTW3 Ultra Gaming, 8GB GDDR6, iCX3 Technology, ARGB LED, Metal Backplate",
				Price: 999.99},
			URL: "https://www.amazon.com/EVGA-08G-P5-3767-KR-GeForce-Technology-Backplate/dp/B08L8L9TCZ"},
	}

	type Test struct {
		search      string
		results     []search.Result
		expectedUrl string
	}

	tests := []Test{
		{
			search:      "rtx 3070",
			results:     searchResults,
			expectedUrl: "https://www.amazon.com/MSI-GeForce-256-Bit-Architecture-Graphics/dp/B08KWLMZV4",
		},
	}

	for _, test := range tests {
		testname := fmt.Sprintf(
			"%s",
			test.search,
		)
		t.Run(testname, func(t *testing.T) {
			search.SortByPrice(test.results)
			cheaperResult := test.results[0]
			res := cheaperResult.URL

			if res != test.expectedUrl {
				t.Errorf("got url %s; want %s", res, test.expectedUrl)
				prettyPrintResultList("results", test.results, t)
			}
		})
	}
}

func prettyPrintResultList(title string, results []search.Result, t *testing.T) {
	t.Errorf("%s: \n", title)
	for _, result := range results {
		t.Errorf("%+v\n", result)
	}
}
