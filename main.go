package main

import (
	"fmt"
	"time"
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
func login(w http.ResponseWriter, r *http.Request) {

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, ok := users[username]
	if !ok || !checkPasswodHash(password, user.HashedPassword) {
		er:= http.StatusNotFound
		http.Error(w, "User not found", er)
		return
	}

	sessionToken := generateToken(32)
	csrfToken := generateToken(32)

	//Set Session Cookie
	http.SetCookie(w, &http.Cookie{
		Name: "session_token",
		Value: sessionToken,
		Expires: time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	//Set CSRF Cookie in a cookie
	http.SetCookie(w, &http.Cookie{
		Name: "csrf_token",
		Value: csrfToken,
		Expires: time.Now().Add(24 * time.Hour),
		HttpOnly: false,
	})


	//Store token in the database
	user.SessionToken = sessionToken
	user.CSRFToken = csrfToken
	users[username] = user
	

	fmt.Fprintln(w, "Login successful")

}
func logout(w http.ResponseWriter, r *http.Request) {}
func protected(w http.ResponseWriter, r *http.Request) {}
