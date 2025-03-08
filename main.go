package main

import (
	"fmt"
	
	"log"
	"net/http"
	
)

type Login struct {

	HashedPassword string
	SessionToken   string
	CSRFToken      string

	}

	var users = map[string]Login{}

const (
	port = ":8080"
)

func main() {

	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/protected", protected)


	// db, err := InitDB()
	// if err != nil {
	// 	log.Fatalf("Failed to initialize database: %v", err)
	// }
	// defer db.Close()

	// fmt.Println("Database setup completed successfully.")

	// Start the HTTP server on port 8080 and log any fatal errors.
	fmt.Println("Server started on port 8080")
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

func register(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid Method", er)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	if len(username)< 8 || len(password) < 8 {
		er:= http.StatusNotAcceptable
		http.Error(w, "Username and Password must be at least 8 characters long", er)
		return
	}

	if _, ok := users[username]; ok {
		er:= http.StatusConflict
		http.Error(w, "Username already exists", er)
		return
	}

	hashedPassword, _ := hashPassword(password)
	users[username] = Login{
		HashedPassword: hashedPassword,
	}

	fmt.Fprintln(w,"User registered successfully.")
}
func login(w http.ResponseWriter, r *http.Request) {}
func logout(w http.ResponseWriter, r *http.Request) {}
func protected(w http.ResponseWriter, r *http.Request) {}
