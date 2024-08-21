package passkey

import (
	"github.com/go-webauthn/webauthn/webauthn"
)

var (
	WebAuthn  *webauthn.WebAuthn
	Datastore PasskeyStore
)
