package handlers

import (
	"database/sql"
	"door-control/db"
	"door-control/models"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/sessions"
)

type RegisterHandler struct {
	DB        *db.DB
	WebAuthn  *webauthn.WebAuthn
	Store     *sessions.CookieStore
	Templates *template.Template
}

func (h *RegisterHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	h.Templates.ExecuteTemplate(w, "register.html", nil)
}

func (h *RegisterHandler) BeginRegistration(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	displayName := r.FormValue("displayName")

	if username == "" || displayName == "" {
		http.Error(w, "Username and display name required", http.StatusBadRequest)
		return
	}

	userID, _, err := h.DB.GetUserByUsername(username)
	if err == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error checking user: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	userID, err = h.DB.CreateUser(username, displayName)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	user := models.User{
		ID:          userID,
		Username:    username,
		DisplayName: displayName,
		Credentials: []webauthn.Credential{},
		DB:          h.DB,
	}

	options, session, err := h.WebAuthn.BeginRegistration(user)
	if err != nil {
		log.Printf("Error beginning registration: %v", err)
		http.Error(w, "Failed to begin registration", http.StatusInternalServerError)
		return
	}

	sess, _ := h.Store.Get(r, "webauthn-session")
	sessionData, _ := json.Marshal(session)
	sess.Values["registration"] = sessionData
	sess.Values["userID"] = userID
	sess.Save(r, w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(options)
}

func (h *RegisterHandler) FinishRegistration(w http.ResponseWriter, r *http.Request) {
	sess, err := h.Store.Get(r, "webauthn-session")
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	sessionData, ok := sess.Values["registration"].([]byte)
	if !ok {
		http.Error(w, "No registration in progress", http.StatusBadRequest)
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

	user := models.User{
		ID:          userID,
		Username:    string(sessionDataStruct.UserID),
		DisplayName: displayName,
		Credentials: []webauthn.Credential{},
		DB:          h.DB,
	}

	credential, err := h.WebAuthn.FinishRegistration(user, sessionDataStruct, r)
	if err != nil {
		log.Printf("Error finishing registration: %v", err)
		http.Error(w, "Failed to finish registration", http.StatusInternalServerError)
		return
	}

	if err := h.DB.SaveCredential(userID, credential.ID, credential.PublicKey); err != nil {
		log.Printf("Error saving credential: %v", err)
		http.Error(w, "Failed to save credential", http.StatusInternalServerError)
		return
	}

	delete(sess.Values, "registration")
	sess.Values["authenticated"] = true
	sess.Values["userID"] = userID
	sess.Save(r, w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
