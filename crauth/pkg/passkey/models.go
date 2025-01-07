package passkey

import (
	"encoding/json"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type PasskeyUser interface {
	webauthn.User
	AddCredential(*webauthn.Credential)
	ReplaceCredential(*webauthn.Credential)
	UpdateCredential(*webauthn.Credential)
	GetUserCredsJson() string
	SetCredential([]webauthn.Credential)
	CredentialExcludeList() []protocol.CredentialDescriptor
}

type PasskeyStore interface {
	GetUser(userName string) PasskeyUser
	CheckExistUser(userName string) bool
	CheckHasUser() bool
	SaveUser(PasskeyUser)
	GetSession(token string) webauthn.SessionData
	SaveSession(token string, data webauthn.SessionData)
	DeleteSession(token string)
	RemoveUser(userKey string)
	InsertPasskeyUser(username string, passKeyUser User)
	GetAllUsername() []string
}

type User struct {
	ID       []byte
	Username string
	creds    []webauthn.Credential
}

func (o *User) WebAuthnID() []byte {
	return o.ID
}

func (o *User) WebAuthnName() string {
	return o.Username
}

func (o *User) WebAuthnDisplayName() string {
	return o.Username
}

func (o *User) WebAuthnIcon() string {
	return "/static/images/user-icon.svg"
}

func (o *User) WebAuthnCredentials() []webauthn.Credential {
	return o.creds
}

func (o *User) AddCredential(credential *webauthn.Credential) {
	o.creds = append(o.creds, *credential)
}

func (o *User) ReplaceCredential(credential *webauthn.Credential) {
	o.creds = []webauthn.Credential{*credential}
}

func (o *User) UpdateCredential(credential *webauthn.Credential) {
	for i, c := range o.creds {
		if string(c.ID) == string(credential.ID) {
			o.creds[i] = *credential
		}
	}
}

func (o *User) SetCredential(credential []webauthn.Credential) {
	o.creds = credential
}

func (o *User) GetUserCredsJson() string {
	if o.creds == nil || len(o.creds) == 0 {
		return ""
	}
	jsonByte, err := json.Marshal(o.creds)
	if err != nil {
		return ""
	}
	return string(jsonByte)
}

func (o *User) CredentialExcludeList() []protocol.CredentialDescriptor {
	credentialExcludeList := []protocol.CredentialDescriptor{}
	for _, cred := range o.creds {
		descriptor := protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: cred.ID,
		}
		credentialExcludeList = append(credentialExcludeList, descriptor)
	}
	return credentialExcludeList
}
