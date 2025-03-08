package main

import (
	"fmt"
	"log"
	"net/http"

	"gitea.com/ypatios/forum/database"
	"gitea.com/ypatios/forum/server"
)

const (
	port = ":8080"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	fmt.Println("Database setup completed successfully.")

	server.NewServer(db).Start(port)
	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Debug route works!"))
	})
}
