package main

import (
	"errors"
	"log"
	"os"

	"github.com/nikoksr/konfetty"
	"github.com/nikoksr/konfetty/examples"
)

type PizzaConfig struct {
	Size        string
	Crust       string
	Sauce       string
	Cheese      string
	Toppings    []string
	ExtraCheese bool
}

type OrderConfig struct {
	Pizza    PizzaConfig
	Quantity int
	Delivery bool
	Address  string
}

func main() {
	// Create a basic pizza order. Typically, you'd populate this from a config file or some other source.
	cfg := &OrderConfig{
		Pizza: PizzaConfig{
			Size:  "medium",
			Crust: "thin",
		},
		Quantity: 1,
		Address:  "A-123, 4th Street, New York",
	}

	// Use konfetty to process the config
	cfg, err := konfetty.FromStruct(cfg).
		WithDefaults(
			OrderConfig{
				Pizza: PizzaConfig{
					Sauce:       "tomato",
					Cheese:      "mozzarella",
					Toppings:    []string{"mushrooms"},
					ExtraCheese: false,
				},
			},
		).
		WithTransformer(func(c *OrderConfig) {
			if c.Address != "" {
				c.Delivery = true
			}
		}).
		WithValidator(func(c *OrderConfig) error {
			if c.Delivery && c.Address == "" {
				return errors.New("delivery address is required for delivery orders")
			}
			return nil
		}).
		Build()
	if err != nil {
		log.Fatalf("Error processing pizza order: %v", err)
	}

	// Use the final config as needed...

	examples.PrettyPrint(os.Stdout, cfg)

	// The final config would look like this:
	//
	// {
	//   "Pizza": {
	//     "Size": "medium",           // Kept original value
	//     "Crust": "thin",            // Kept original value
	//     "Sauce": "tomato",          // Applied from defaults
	//     "Cheese": "mozzarella",     // Applied from defaults
	//     "Toppings": [
	//       "mushrooms"
	//     ],                          // Applied from defaults
	//     "ExtraCheese": false        // Applied from defaults
	//   },
	//   "Quantity": 1,                // Kept original value
	//   "Delivery": true,             // Applied from transformer
	//   "Address": "A-123, 4th Street, New York"  // Kept original value
	// }
}
