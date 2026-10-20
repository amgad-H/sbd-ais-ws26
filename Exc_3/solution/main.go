package main

import (
	"embed"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"ordersystem/repository"
	"ordersystem/rest"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "ordersystem/docs"
	// OpenApi
	httpSwagger "github.com/swaggo/http-swagger"
)

// embeddedFrontend embeds the frontend into the binary
//
//go:embed frontend/*
var embeddedFrontend embed.FS

// @title				Order System
// @description			This system enables drink orders and should not be used for the forbidden Hungover Games.
// @contact.name		Your Name
func main() {
	// The database is created once and passed to every handler,
	// so all routes work on the same data.
	db := repository.NewDatabaseHandler()

	// The router decides which handler is called for which URL and HTTP method.
	r := chi.NewRouter()
	// The Logger middleware prints every incoming request to the console,
	// which is useful for seeing what the server is doing.
	r.Use(middleware.Logger)

	// Render Frontend
	// The frontend files are embedded into the binary at compile time,
	// so the server can be run without shipping the frontend folder separately.
	staticFS, err := fs.Sub(embeddedFrontend, "frontend")
	if err != nil {
		log.Fatal(err)
	}
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, staticFS, "index.html")
	})
	// Menu Routes
	// Each route connects a method and a path to a handler from the rest package.
	// The handlers receive the database here, once, when the routes are set up.
	r.Get("/api/menu", rest.GetMenu(db))
	// Order Routes
	r.Get("/api/order/all", rest.GetOrders(db))
	r.Get("/api/order/totalled", rest.GetOrdersTotal(db))
	r.Post("/api/order", rest.PostOrder(db))
	// OpenAPI Routes
	// Serves the generated Swagger UI, where the API can be explored and tested.
	r.Get("/openapi/*", httpSwagger.WrapHandler)

	slog.Info("⚡⚡⚡ Order System is up and running ⚡⚡⚡")
	// ListenAndServe blocks and handles requests until the program is stopped.
	err = http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatal(err)
	}
}
