// Package repository is responsible for storing and retrieving data.
//
// The rest of the application only talks to the data through the methods
// in this package. Because of that, the in-memory storage used here could be
// replaced with a real database later without changing the HTTP handlers.
package repository

import (
	"ordersystem/model"
	"time"
)

// DatabaseHandler is a simple in-memory database.
// All data is held in slices, so it is lost when the server stops.
type DatabaseHandler struct {
	// drinks represent all available drinks
	drinks []model.Drink
	// orders serves as order history
	orders []model.Order
}

// NewDatabaseHandler creates a new database pre-filled with some test data,
// so the API and the dashboard have something to show right after startup.
// It returns a pointer so every part of the application shares the same
// database instead of working on separate copies.
func NewDatabaseHandler() *DatabaseHandler {
	drinks := []model.Drink{
		{ID: 1, Name: "Beer", Price: 3.50, Description: "Cold and refreshing"},
		{ID: 2, Name: "Spritzer", Price: 2.80, Description: "White wine with soda water"},
		{ID: 3, Name: "Coffee", Price: 2.20, Description: "For the morning after"},
	}

	// The test orders are spread over the last few minutes so the
	// "Order History" chart on the dashboard shows a line over time.
	orders := []model.Order{
		{DrinkID: 1, Amount: 2, CreatedAt: time.Now().Add(-10 * time.Minute)},
		{DrinkID: 2, Amount: 1, CreatedAt: time.Now().Add(-5 * time.Minute)},
		{DrinkID: 1, Amount: 3, CreatedAt: time.Now()},
	}

	return &DatabaseHandler{
		drinks: drinks,
		orders: orders,
	}
}

// GetDrinks returns all drinks on the menu.
func (db *DatabaseHandler) GetDrinks() []model.Drink {
	return db.drinks
}

// GetOrders returns the complete order history.
func (db *DatabaseHandler) GetOrders() []model.Order {
	return db.orders
}

// GetTotalledOrders sums up the ordered amount per drink.
// The key of the returned map is the DrinkID and the value is the total
// amount ordered for that drink.
//
// For example, the orders {drink 1, amount 2}, {drink 2, amount 1} and
// {drink 1, amount 3} result in the map {1: 5, 2: 1}.
func (db *DatabaseHandler) GetTotalledOrders() map[uint64]uint64 {
	// The map has to be created with make. A map that is only declared
	// is nil, and writing to a nil map causes a panic.
	totalledOrders := make(map[uint64]uint64)

	for _, order := range db.orders {
		// Reading a key that does not exist yet returns 0, so the first
		// order of each drink is handled without any extra check.
		totalledOrders[order.DrinkID] += order.Amount
	}

	return totalledOrders
}

// AddOrder appends a new order to the order history.
//
// The method uses a pointer receiver (*DatabaseHandler) so the change is
// made to the shared database rather than to a copy of it.
func (db *DatabaseHandler) AddOrder(order *model.Order) {
	// append returns a new slice, so the result has to be assigned back
	// to db.orders. The order is dereferenced (*order) because the slice
	// stores Order values, not pointers.
	db.orders = append(db.orders, *order)
}
