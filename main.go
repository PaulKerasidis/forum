package main

import (
	"fmt"
	"log"
	"net/http"

)

const (
	port = ":8080"
)

func main() {
	db, err := InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	fmt.Println("Database setup completed successfully.")

	// Start the HTTP server on port 8080 and log any fatal errors.
	fmt.Println("Server started on port 8080")
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
