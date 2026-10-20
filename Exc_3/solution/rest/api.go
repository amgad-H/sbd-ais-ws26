// Package rest contains the HTTP handlers of the order system.
//
// The handlers translate between HTTP and the repository: they read the
// request, call the matching database method and send the result back as JSON.
//
// Every handler is a function that takes the database and returns an
// http.HandlerFunc. Chi requires handlers to have the exact signature
// func(w http.ResponseWriter, r *http.Request), which leaves no room for
// extra parameters. The outer function receives the database once when the
// routes are registered in main.go, and the returned inner function keeps
// access to it (a closure). This avoids the need for a global variable.
//
// The comment blocks starting with @ above each handler are read by swag
// to generate the OpenAPI documentation.
package rest

import (
	"encoding/json"
	"net/http"
	"ordersystem/model"
	"ordersystem/repository"
	"time"

	"github.com/go-chi/render"
)

// GetMenu 			godoc
// @tags 			Menu
// @Description 	Returns the menu of all drinks
// @Produce  		json
// @Success 		200 {array} model.Drink
// @Router 			/api/menu [get]
func GetMenu(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		drinks := db.GetDrinks()
		render.Status(r, http.StatusOK)
		// render.JSON encodes the slice as a JSON array and sets the
		// Content-Type header to application/json.
		render.JSON(w, r, drinks)
	}
}

// GetOrders 		godoc
// @tags 			Order
// @Description 	Returns all orders
// @Produce  		json
// @Success 		200 {array} model.Order
// @Router 			/api/order/all [get]
func GetOrders(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orders := db.GetOrders()
		render.Status(r, http.StatusOK)
		render.JSON(w, r, orders)
	}
}

// GetOrdersTotal 	godoc
// @tags 			Order
// @Description 	Returns how many of each drink have been ordered (drink id -> amount)
// @Produce  		json
// @Success 		200 {object} map[string]int
// @Router 			/api/order/totalled [get]
func GetOrdersTotal(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		totals := db.GetTotalledOrders()
		render.Status(r, http.StatusOK)
		// A map is encoded as a JSON object. JSON keys are always strings,
		// so the drink IDs appear as "1", "2", ... in the response.
		render.JSON(w, r, totals)
	}
}

// PostOrder 		godoc
// @tags 			Order
// @Description 	Adds an order to the db
// @Accept 			json
// @Param 			b body model.Order true "Order"
// @Produce  		json
// @Success 		200
// @Failure     	400
// @Router 			/api/order [post]
func PostOrder(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var order model.Order

		// Decode the JSON request body into the order. A pointer to the
		// order is passed so the decoder can fill in its fields.
		err := json.NewDecoder(r.Body).Decode(&order)
		if err != nil {
			// The request body is not valid JSON, so the client receives
			// 400 Bad Request. The return stops the handler here so the
			// invalid order is not stored.
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, "invalid order: "+err.Error())
			return
		}

		// If the client did not send a timestamp, the current time is used.
		if order.CreatedAt.IsZero() {
			order.CreatedAt = time.Now()
		}

		db.AddOrder(&order)

		render.Status(r, http.StatusOK)
		render.JSON(w, r, "ok")
	}
}
