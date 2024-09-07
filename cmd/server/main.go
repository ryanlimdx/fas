// The entry point; This file sets up the server and routes for the application.
package main

import (
	"fas/internal/database"
	"fas/internal/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	// Connect to database
	db, err := database.SetupDB()
	if err != nil {
		log.Fatalf("Could not set up database: %v", err)
	}
	defer db.Close()

	// Initialise router
	r := mux.NewRouter()
	
	// Routes (API Endpoints)
	r.HandleFunc("/api/applicants", handlers.CreateApplicant(db)).Methods(http.MethodPost)
	r.HandleFunc("/api/applicants", handlers.GetApplicants(db)).Methods(http.MethodGet)
	
	r.HandleFunc("/api/schemes", handlers.CreateScheme(db)).Methods(http.MethodPost)
	r.HandleFunc("/api/schemes", handlers.GetSchemes(db)).Methods(http.MethodGet)
	
	r.HandleFunc("/api/applications", handlers.CreateApplication(db)).Methods(http.MethodPost)
	r.HandleFunc("/api/applications", handlers.GetApplications(db)).Methods(http.MethodGet)

	// r.HandleFunc("/api/schemes/eligible", handlers.GetEligibleSchemes(db)).Methods(http.MethodGet)
	
	// Start server
	log.Fatal(http.ListenAndServe(":8080", r))
}