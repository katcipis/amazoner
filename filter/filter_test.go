package filter

import (
	"fmt"
	"testing"

	"github.com/katcipis/amazoner/search"
)

func TestFilter(t *testing.T) {
	searchResults := []search.Result{
		search.Result{
			Product: search.Product{
				Name:  "MSI GeForce RTX 3070 Ventus 3X OC Gaming Video Card, 8GB GDDR6, PCIe 4.0, 8K, VR Ready, Ray Tracing, 1x HDMI 2.1, 3X DisplayPort 1.4, Triple Fans, HDCP, Mytrix HDMI 2.1 8K Cable",
				Price: 939.99},
			URL: "https://www.amazon.com/MSI-RTX-3070-HDMI-DisplayPort/dp/B08MVFMN35"},
	}

	type Test struct {
		search      string
		maxPrice    uint
		results     []search.Result
		expectedUrl string
	}

	tests := []Test{
		{
			search:      "rtx 3070",
			maxPrice:    1500,
			results:     searchResults,
			expectedUrl: "https://www.amazon.com/MSI-RTX-3070-HDMI-DisplayPort/dp/B08MVFMN35",
		},
	}

	for _, test := range tests {
		testname := fmt.Sprintf(
			"%sMax%d",
			test.search,
			test.maxPrice,
		)
		t.Run(testname, func(t *testing.T) {
			res, err := Do(test.search, test.maxPrice, test.results)

			if res != test.expectedUrl {
				t.Errorf("got url %s; want %s", res, test.expectedUrl)
				t.Errorf("results:%v", test.results)
				if err != nil {
					t.Errorf("errors:%v", err)
				}
			}
		})
	}
}

// {Product:{Name:MSI GeForce RTX 3070 Ventus 3X OC Gaming Video Card, 8GB GDDR6, PCIe 4.0, 8K, VR Ready, Ray Tracing, 1x HDMI 2.1, 3X DisplayPort 1.4, Triple Fans, HDCP, Mytrix HDMI 2.1 8K Cable Price:939.99} URL:https://www.amazon.com/MSI-RTX-3070-HDMI-DisplayPort/dp/B08MVFMN35}
// {Product:{Name:MSI GeForce RTX 3070 Ventus 3X OC Gaming Video Card, 8GB GDDR6, PCIe 4.0, 8K, VR Ready, Ray Tracing, 1x HDMI 2.1, 3X DisplayPort 1.4, Triple Fans, HDCP, Battlefield V Price:939.99} URL:https://www.amazon.com/MSI-RTX-3070-DisplayPort-Battlefield/dp/B08MVHD9Z9}
// {Product:{Name:CyberpowerPC Gamer Xtreme VR Gaming PC, Intel i5-10400F 2.9GHz, GeForce GTX 1660 Super 6GB, 8GB DDR4, 500GB NVMe SSD, WiFi Ready & Win 10 Home (GXiVR8060A10) Price:799.99} URL:https://www.amazon.com/CyberpowerPC-Xtreme-i5-10400F-GeForce-GXiVR8060A10/dp/B08FBK2DK5}
// {Product:{Name:MSI Gaming GeForce RTX 3070 8GB GDRR6 256-Bit HDMI/DP Tri-Frozr 2 TORX Fan 4.0 Ampere Architecture RGB OC Graphics Card (RTX 3070 Gaming X Trio) Price:959} URL:https://www.amazon.com/MSI-GeForce-RTX-3070-Architecture/dp/B08KWN2LZG}
// {Product:{Name:MSI Gaming GeForce RTX 3070 8GB GDRR6 256-Bit HDMI/DP TORX Fan 3.0 Ampere Architecture OC Graphics Card (RTX 3070 Ventus 3X OC) Price:909.99} URL:https://www.amazon.com/MSI-GeForce-256-Bit-Architecture-Graphics/dp/B08KWLMZV4}
// {Product:{Name:PNY GeForce RTX 3070 8GB XLR8 Gaming Revel Epic-X RGB Triple Fan Graphics Card Price:949.99} URL:https://www.amazon.com/PNY-GeForce-Gaming-Epic-X-Graphics/dp/B08HBJB7YD}
// {Product:{Name:ARESGAME 750W Power Supply Semi Modular 80+ Bronze PSU (AGV750) Price:79.99} URL:https://www.amazon.com/ARESGAME-Supply-Modular-Bronze-AGV750/dp/B08JM12SQ5}
// {Product:{Name:MSI GeForce RTX 3070 Ventus 3X OC Gaming Video Card, 8GB GDDR6, PCIe 4.0, 8K, VR Ready, Ray Tracing, 1x HDMI 2.1, 3X DisplayPort 1.4, Triple Fans, HDCP, Mytrix HDMI 2.1 8K Cable, Battlefield V Price:939.99} URL:https://www.amazon.com/MSI-RTX-3070-DisplayPort-Battlefield/dp/B08MVH3QJF}
// {Product:{Name:Beelink U57 Mini PC with Intel Core i5-5257u Processor(up to 3.10 GHz)&Windows 10 Pro,8G DDR3L/256G SSD High Performance Business Mini Computer,2.4G/5G Dual WiFi,BT4.2,Dual HDMI Ports Price:379} URL:https://www.amazon.com/Beelink-U57-Processor-256G-Performance/dp/B0879KKTCB}
// {Product:{Name:EVGA 08G-P5-3767-KR GeForce RTX 3070 FTW3 Ultra Gaming, 8GB GDDR6, iCX3 Technology, ARGB LED, Metal Backplate Price:999.99} URL:https://www.amazon.com/EVGA-08G-P5-3767-KR-GeForce-Technology-Backplate/dp/B08L8L9TCZ}
