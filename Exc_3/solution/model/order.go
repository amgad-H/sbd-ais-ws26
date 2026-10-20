package model

import "time"

// Order represents a single purchase of one type of drink.
// Several units of the same drink can be ordered at once using Amount.
type Order struct {
	// DrinkID is a foreign key that references Drink.ID.
	DrinkID uint64 `json:"drink_id"`
	// CreatedAt is the time the order was placed. In JSON it is encoded
	// as an RFC 3339 timestamp, e.g. "2026-10-20T10:30:00+02:00".
	CreatedAt time.Time `json:"created_at"`
	// Amount is the number of units ordered. It is unsigned because
	// a negative amount would not make sense.
	Amount uint64 `json:"amount"`
}
