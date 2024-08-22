package passkey

import (
	"crauth/pkg/logpack"
	"crauth/pkg/utils"
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"
)

func NewInMem() *InMem {
	return &InMem{
		users:    make(map[string]PasskeyUser),
		sessions: make(map[string]webauthn.SessionData),
	}
}

type InMem struct {
	users    map[string]PasskeyUser
	sessions map[string]webauthn.SessionData
}

func (i *InMem) GetSession(token string) webauthn.SessionData {
	logpack.Info(fmt.Sprintf("[DEBUG] GetSession: %v", i.sessions[token]), utils.GetFuncName())
	return i.sessions[token]
}

func (i *InMem) SaveSession(token string, data webauthn.SessionData) {
	logpack.Info(fmt.Sprintf("[DEBUG] SaveSession: %s - %v", token, data), utils.GetFuncName())
	i.sessions[token] = data
}

func (i *InMem) DeleteSession(token string) {
	logpack.Info(fmt.Sprintf("[DEBUG] DeleteSession: %v", token), utils.GetFuncName())
	delete(i.sessions, token)
}

func (i *InMem) CheckExistUser(userKey string) bool {
	_, ok := i.users[userKey]
	return ok
}

func (i *InMem) CheckHasUser() bool {
	return len(i.users) > 0
}

func (i *InMem) GetAllUsername() []string {
	list := make([]string, 0)
	for key := range i.users {
		list = append(list, key)
	}
	return list
}

func (i *InMem) GetUser(userName string) PasskeyUser {
	logpack.Info(fmt.Sprintf("[DEBUG] GetUser: %v", userName), utils.GetFuncName())
	if _, ok := i.users[userName]; !ok {
		logpack.Info(fmt.Sprintf("[DEBUG] GetUser: creating new user: %v", userName), utils.GetFuncName())
		i.users[userName] = &User{
			ID:       []byte(userName),
			Username: userName,
		}
	}

	return i.users[userName]
}

func (i *InMem) InsertPasskeyUser(username string, passKeyUser User) {
	i.users[username] = &passKeyUser
}

func (i *InMem) RemoveUser(userKey string) {
	if i.users == nil {
		return
	}
	removeKey := ""
	for key, value := range i.users {
		if value.WebAuthnName() == userKey {
			removeKey = key
			break
		}
	}
	delete(i.users, removeKey)
}

func (i *InMem) SaveUser(user PasskeyUser) {
	logpack.Info(fmt.Sprintf("[DEBUG] SaveUser: %v", user.WebAuthnName()), utils.GetFuncName())
	i.users[user.WebAuthnName()] = user
}
