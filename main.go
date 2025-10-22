package main

import (
	"door-control/db"
	"door-control/handlers"
	"html/template"
	"log"
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/sessions"
)

func main() {
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

	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	registerHandler := &handlers.RegisterHandler{
		DB:        database,
		WebAuthn:  webAuthn,
		Store:     store,
		Templates: tmpl,
	}

	loginHandler := &handlers.LoginHandler{
		DB:        database,
		WebAuthn:  webAuthn,
		Store:     store,
		Templates: tmpl,
	}

	dashboardHandler := &handlers.DashboardHandler{
		DB:        database,
		Store:     store,
		Templates: tmpl,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	http.HandleFunc("/register", registerHandler.RegisterPage)
	http.HandleFunc("/register/begin", registerHandler.BeginRegistration)
	http.HandleFunc("/register/finish", registerHandler.FinishRegistration)

	http.HandleFunc("/login", loginHandler.LoginPage)
	http.HandleFunc("/login/begin", loginHandler.BeginLogin)
	http.HandleFunc("/login/finish", loginHandler.FinishLogin)
	http.HandleFunc("/logout", loginHandler.Logout)

	http.HandleFunc("/dashboard", dashboardHandler.Dashboard)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server starting on :8080")
	log.Println("Production URL: https://doorctrl.sooth.dev")
	log.Println("Local access: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
