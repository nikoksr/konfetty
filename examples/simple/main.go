package main

import (
	"errors"
	"log"

	"github.com/nikoksr/konfetty"
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
		Delivery: true,
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
				Delivery: true,
			},
		).
		WithTransformer(func(c *OrderConfig) {
			if c.Pizza.Size == "large" && !c.Pizza.ExtraCheese {
				c.Pizza.ExtraCheese = true
				c.Pizza.Toppings = append(c.Pizza.Toppings, "extra cheese")
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

	// The final config would look like this:
	//
	// 	OrderConfig{
	//   Pizza: PizzaConfig{
	//     Size:     "medium",
	//     Crust:    "thin",
	//     Sauce:    "tomato",
	//     Cheese:   "mozzarella",
	//     Toppings: []string{
	//       "mushrooms",
	//     },
	//     ExtraCheese: false,
	//   },
	//   Quantity: 1,
	//   Delivery: true,
	//   Address:  "A-123, 4th Street, New York",
	// }
}
