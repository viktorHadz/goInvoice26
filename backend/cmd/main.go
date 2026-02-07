package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app/clients"
	"github.com/viktorHadz/goInvoice26/internal/config"
	"github.com/viktorHadz/goInvoice26/internal/db"
)

func main() {
	// Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// DB
	ctx := context.Background()

	dbConn, err := db.OpenDB(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	if err := db.Migrate(ctx, dbConn); err != nil {
		log.Fatal(err)
	}

	// Build deps
	q := clients.ClientQueries{DB: dbConn}
	svc := clients.ClientService{Q: q}
	clientAPI := clients.ClientAPI{Svc: svc}

	// Server
	r := gin.Default()
	clientAPI.Register(r)

	// Create a server and mux
	port := ":" + cfg.Port
	router := chi.NewRouter()

	// Register routes
	clientsRouter(router)
	http.ListenAndServe(port, router)

	log.Printf("env=%s db=%s", cfg.Env, cfg.DBPath)
	log.Printf("API listening on :%s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}

// Registers a clients router with POST, GET, PATCH, DELETE functionality
func clientsRouter(r chi.Router) {
	r.Route("/clients", func(r chi.Router) {
		r.Post("/", createClient) // CREATE  POST /clients
		r.Get("/", listClients)   // READ    GET  /clients

		r.Route("/{id}", func(r chi.Router) {
			r.Patch("/", updateClient)  // UPDATE  PATCH /clients/{id}
			r.Delete("/", deleteClient) // DELETE DELETE /clients/{id}
		})
	})
}

func createClient(w http.ResponseWriter, r *http.Request) {
	// DB call here
}
func listClients(w http.ResponseWriter, r *http.Request) {
	// DB call here
}
func updateClient(w http.ResponseWriter, r *http.Request) {
	// DB call here
}
func deleteClient(w http.ResponseWriter, r *http.Request) {
	// DB call here
}
