package models

import (
	"door-control/internal/db"

	"github.com/go-webauthn/webauthn/webauthn"
)

type User struct {
	ID          int64
	Username    string
	DisplayName string
	Credentials []webauthn.Credential
	DB          *db.DB
}

func (u User) WebAuthnID() []byte {
	return []byte(u.Username)
}

func (u User) WebAuthnName() string {
	return u.Username
}

func (u User) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u User) WebAuthnIcon() string {
	return ""
}

func (u User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

func LoadUserCredentials(database *db.DB, userID int64) ([]webauthn.Credential, error) {
	credIDs, err := database.GetCredentialsByUserID(userID)
	if err != nil {
		return nil, err
	}

	var credentials []webauthn.Credential
	for _, credID := range credIDs {
		_, publicKey, signCount, backupEligible, backupState, err := database.GetCredential(credID)
		if err != nil {
			continue
		}

		credentials = append(credentials, webauthn.Credential{
			ID:              credID,
			PublicKey:       publicKey,
			AttestationType: "none",
			Flags: webauthn.CredentialFlags{
				BackupEligible: backupEligible,
				BackupState:    backupState,
			},
			Authenticator: webauthn.Authenticator{
				SignCount: uint32(signCount),
			},
		})
	}

	return credentials, nil
}
