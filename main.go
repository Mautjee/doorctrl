package main

import (
	"door-control/internal/db"
	"door-control/internal/routes"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			return parsed
		}
	}
	return defaultVal
}

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

	sessionSecret := []byte("super-secret-key-change-in-production")
	if envSecret := []byte(os.Getenv("SESSION_SECRET")); len(envSecret) > 0 {
		sessionSecret = envSecret
	} else {
		log.Println("WARNING: Using default SESSION_SECRET. Set SESSION_SECRET environment variable in production!")
	}
	store := sessions.NewCookieStore(sessionSecret)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	routes.Setup(database, webAuthn, store, tmpl)

	log.Println("========================================")
	log.Println("Door Control System Starting")
	log.Println("========================================")
	log.Println("Server listening on :8080")
	log.Println("Production URL: https://doorctrl.sooth.dev")
	log.Println("Local access: http://localhost:8080")
	log.Printf("WebAuthn RPID: %s", wconfig.RPID)
	log.Printf("Rate limiting: 5 requests per second per IP")
	log.Printf("Studio location: %.6f, %.6f",
		getEnvFloat("STUDIO_LATITUDE", 0),
		getEnvFloat("STUDIO_LONGITUDE", 0))
	log.Println("========================================")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
