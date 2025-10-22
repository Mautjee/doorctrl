package routes

import (
	"door-control/internal/db"
	"door-control/internal/handlers"
	"html/template"
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/sessions"
)

func Setup(database *db.DB, webAuthn *webauthn.WebAuthn, store *sessions.CookieStore, tmpl *template.Template) {
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

	bookingHandler := &handlers.BookingHandler{
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

	http.HandleFunc("/booking", bookingHandler.BookingPage)
	http.HandleFunc("/booking/create", bookingHandler.CreateBooking)
	http.HandleFunc("/bookings", bookingHandler.GetUserBookings)
	http.HandleFunc("/unlock", bookingHandler.UnlockDoor)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}
