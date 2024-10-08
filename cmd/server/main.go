// The entry point. This file sets up the server and routes for the application.
package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"fas/internal/database"
	"fas/internal/handlers"
	"fas/internal/middleware"
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
	// Applicants
	r.Handle("/api/applicants", middleware.ValidateApplicant(handlers.CreateApplicant(db))).Methods(http.MethodPost)
	r.Handle("/api/applicants/{id}", middleware.ValidateApplicant(handlers.UpdateApplicant(db))).Methods(http.MethodPut)
	r.HandleFunc("/api/applicants", handlers.GetApplicants(db)).Methods(http.MethodGet)
	r.HandleFunc("/api/applicants/{id}", handlers.DeleteApplicant(db)).Methods(http.MethodDelete)
	
	// Schemes
	r.Handle("/api/schemes", middleware.ValidateScheme(handlers.CreateScheme(db))).Methods(http.MethodPost)
	r.Handle("/api/schemes/{id}", middleware.ValidateScheme(handlers.UpdateScheme(db))).Methods(http.MethodPut)
	r.HandleFunc("/api/schemes", handlers.GetSchemes(db)).Methods(http.MethodGet)
	r.HandleFunc("/api/schemes/eligible", handlers.GetEligibleSchemes(db)).Methods(http.MethodGet)
	r.HandleFunc("/api/schemes/{id}", handlers.DeleteScheme(db)).Methods(http.MethodDelete)
	
	// Applications
	r.HandleFunc("/api/applications", handlers.CreateApplication(db)).Methods(http.MethodPost)
	r.HandleFunc("/api/applications", handlers.GetApplications(db)).Methods(http.MethodGet)
	r.HandleFunc("/api/applications/{id}", handlers.UpdateApplication(db)).Methods(http.MethodPut)
	r.HandleFunc("/api/applications/{id}", handlers.DeleteApplication(db)).Methods(http.MethodDelete)
	
	// Start server
	log.Fatal(http.ListenAndServe(":8080", r))
}