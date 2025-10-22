package main

import (
	"door-control/internal/db"
	"door-control/internal/routes"
	"html/template"
	"log"
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	database, err := db.InitDB("./door-control.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	wconfig := &webauthn.Config{
		RPDisplayName: "Door Control",
		RPID:          "doorctrl.sooth.dev",
		RPOrigins:     []string{"https://doorctrl.sooth.dev", "http://localhost:8080"},
	}

	webAuthn, err := webauthn.New(wconfig)
	if err != nil {
		log.Fatalf("Failed to create WebAuthn: %v", err)
	}

	store := sessions.NewCookieStore([]byte("super-secret-key-change-in-production"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	tmpl := template.Must(template.New("").ParseGlob("templates/**/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/*.html"))

	routes.Setup(database, webAuthn, store, tmpl)

	log.Println("Server starting on :8080")
	log.Println("Production URL: https://doorctrl.sooth.dev")
	log.Println("Local access: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
