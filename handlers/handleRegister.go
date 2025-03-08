package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type LoginUser struct {
	HashedPassword string
	SessionToken   string
	CSRFToken      string
}

// Key is the username
var users = map[string]LoginUser{}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		err := http.StatusMethodNotAllowed
		http.Error(w, "invalid method", err)
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	if len(username) < 8 || len(password) < 8 {
		err := http.StatusNotAcceptable
		http.Error(w, "invalid username or password", err)
		return
	}
	if _, ok := users[username]; ok {
		err := http.StatusConflict
		http.Error(w, "username already exists", err)
		return
	}
	hashedPassword := hashPassword(password)
	users[username] = LoginUser{
		HashedPassword: hashedPassword,
		SessionToken:   "",
		CSRFToken:      "",
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		err := http.StatusMethodNotAllowed
		http.Error(w, "invalid request method", err)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, ok := users[username]

	if !ok || !checkPasswordHash(password, user.HashedPassword) {
		err := http.StatusUnauthorized
		http.Error(w, "invalid username or password", err)
		return
	}

	sessionToken := generateToken(32)

	http.SetCookie(w, &http.Cookie{
		Name:  "session_token",
		Value: sessionToken,
	})

	fmt.Fprintln(w, "Login successful")
}

func hashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes)
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("failed to generate token: %v", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
