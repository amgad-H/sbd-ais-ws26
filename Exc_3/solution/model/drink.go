// Package model defines the data structures used throughout the order system.
// It only describes the shape of the data and contains no logic.
package model

// Drink represents a single item on the menu.
//
// Fields start with a capital letter so they are exported. The JSON encoder
// ignores unexported (lowercase) fields, so they would not appear in responses.
// The struct tags define the snake_case names used in the JSON representation,
// which is what the frontend and API consumers expect.
type Drink struct {
	// ID uniquely identifies the drink and is referenced by Order.DrinkID.
	ID uint64 `json:"id"`
	// Name is the display name shown on the menu.
	Name string `json:"name"`
	// Price is the price per unit. A float is used since prices have decimals.
	Price float64 `json:"price"`
	// Description is a short text shown next to the drink on the menu.
	Description string `json:"description"`
}
