package routes

import (
	"door-control/internal/db"
	"door-control/internal/handlers"
	"door-control/internal/middleware"
	"html/template"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/sessions"
	"golang.org/x/time/rate"
)

func Setup(database *db.DB, webAuthn *webauthn.WebAuthn, store *sessions.CookieStore, tmpl *template.Template) {
	limiter := middleware.NewIPRateLimiter(rate.Every(1*time.Second), 5)

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
	http.HandleFunc("/register/begin", limiter.Limit(registerHandler.BeginRegistration))
	http.HandleFunc("/register/finish", limiter.Limit(registerHandler.FinishRegistration))

	http.HandleFunc("/login", loginHandler.LoginPage)
	http.HandleFunc("/login/begin", limiter.Limit(loginHandler.BeginLogin))
	http.HandleFunc("/login/finish", limiter.Limit(loginHandler.FinishLogin))
	http.HandleFunc("/logout", loginHandler.Logout)

	http.HandleFunc("/dashboard", dashboardHandler.Dashboard)

	http.HandleFunc("/booking", bookingHandler.BookingPage)
	http.HandleFunc("/booking/create", bookingHandler.CreateBooking)
	http.HandleFunc("/bookings", bookingHandler.GetUserBookings)
	http.HandleFunc("/unlock", bookingHandler.UnlockDoor)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 3 {
			ext := path[len(path)-3:]
			if ext == ".js" {
				w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			} else if len(path) > 4 {
				ext4 := path[len(path)-4:]
				if ext4 == ".jpg" || ext4 == "jpeg" {
					w.Header().Set("Content-Type", "image/jpeg")
				} else if ext4 == ".png" {
					w.Header().Set("Content-Type", "image/png")
				} else if ext4 == ".css" {
					w.Header().Set("Content-Type", "text/css; charset=utf-8")
				}
			}
		}
		fs.ServeHTTP(w, r)
	})))
}
