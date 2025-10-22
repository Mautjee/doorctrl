package handlers

import (
	"door-control/internal/db"
	"door-control/internal/models"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/sessions"
)

type LoginHandler struct {
	DB        *db.DB
	WebAuthn  *webauthn.WebAuthn
	Store     *sessions.CookieStore
	Templates *template.Template
}

func (h *LoginHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	h.Templates.ExecuteTemplate(w, "login.html", nil)
}

func (h *LoginHandler) BeginLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")

	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	userID, displayName, err := h.DB.GetUserByUsername(username)
	if err != nil {
		log.Printf("User not found: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	credentials, err := models.LoadUserCredentials(h.DB, userID)
	if err != nil {
		log.Printf("Error loading credentials: %v", err)
		http.Error(w, "Failed to load credentials", http.StatusInternalServerError)
		return
	}

	user := models.User{
		ID:          userID,
		Username:    username,
		DisplayName: displayName,
		Credentials: credentials,
		DB:          h.DB,
	}

	options, session, err := h.WebAuthn.BeginLogin(user)
	if err != nil {
		log.Printf("Error beginning login: %v", err)
		http.Error(w, "Failed to begin login", http.StatusInternalServerError)
		return
	}

	sess, _ := h.Store.Get(r, "webauthn-session")
	sessionData, _ := json.Marshal(session)
	sess.Values["authentication"] = sessionData
	sess.Values["userID"] = userID
	sess.Save(r, w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(options)
}

func (h *LoginHandler) FinishLogin(w http.ResponseWriter, r *http.Request) {
	sess, err := h.Store.Get(r, "webauthn-session")
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	sessionData, ok := sess.Values["authentication"].([]byte)
	if !ok {
		http.Error(w, "No authentication in progress", http.StatusBadRequest)
		return
	}

	userID, ok := sess.Values["userID"].(int64)
	if !ok {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	var sessionDataStruct webauthn.SessionData
	if err := json.Unmarshal(sessionData, &sessionDataStruct); err != nil {
		http.Error(w, "Invalid session data", http.StatusInternalServerError)
		return
	}

	_, displayName, err := h.DB.GetUserByUsername(string(sessionDataStruct.UserID))
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	credentials, err := models.LoadUserCredentials(h.DB, userID)
	if err != nil {
		http.Error(w, "Failed to load credentials", http.StatusInternalServerError)
		return
	}

	user := models.User{
		ID:          userID,
		Username:    string(sessionDataStruct.UserID),
		DisplayName: displayName,
		Credentials: credentials,
		DB:          h.DB,
	}

	credential, err := h.WebAuthn.FinishLogin(user, sessionDataStruct, r)
	if err != nil {
		log.Printf("Error finishing login: %v", err)
		http.Error(w, "Failed to finish login", http.StatusInternalServerError)
		return
	}

	if credential.Authenticator.CloneWarning {
		log.Printf("Clone warning for credential: %v", credential.ID)
	}

	if err := h.DB.UpdateSignCount(credential.ID, int(credential.Authenticator.SignCount)); err != nil {
		log.Printf("Error updating sign count: %v", err)
	}

	delete(sess.Values, "authentication")
	sess.Values["authenticated"] = true
	sess.Values["userID"] = userID
	sess.Values["username"] = string(sessionDataStruct.UserID)
	sess.Save(r, w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *LoginHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := h.Store.Get(r, "webauthn-session")
	sess.Values["authenticated"] = false
	delete(sess.Values, "userID")
	sess.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
