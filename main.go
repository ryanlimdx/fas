// The entry point; This file sets up the server and routes for the application.
package main

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"fas/handlers"
	"fas/database"
)

func main() {

	// Connect to database
	db, err := database.setupDB()
	if err != nil {
		log.Fatal("Could not set up database: %v", err)
	}
	defer db.Close()

	// Initialise router
	r := mux.NewRouter()
	
	// Routes (API Endpoints)
	r.HandleFunc("/api/applicants", handlers.GetApplicants(db)).Methods(http.MethodGet)
	r.HandleFunc("/api/applicants", handlers.CreateApplicant(db)).Methods(http.MethodPost)
	r.HandleFunc("/api/schemes", handlers.GetSchemes(db)).Methods(http.MethodGet)
	r.HandleFunc("/api/schemes/eligible", handlers.GetEligibleSchemes(db)).Methods(http.MethodGet)
	r.HandleFunc("/api/applications", handlers.GetApplications(db)).Methods(http.MethodGet)
	r.HandleFunc("/api/applications", handlers.CreateApplication(db)).Methods(http.MethodPost)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", nil))
}