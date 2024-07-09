package controllers

type AuthController struct {
	BaseController
}

var authUrlBase = ""

func (this *AuthController) BeginRegistration() {
}

func (this *AuthController) FinishRegistration() {
}

func (this *AuthController) AssertionOptions() {
}

func (this *AuthController) AssertionResult() {
}

func (this *AuthController) BeginUpdatePasskey() {
}

func (this *AuthController) FinishUpdatePasskey() {
}

func (this *AuthController) BeginConfirmPasskey() {
}

func (this *AuthController) FinishConfirmPasskey() {
}

func (this *AuthController) CancelRegister() {
}

func (this *AuthController) Quit() {
}

func (this *AuthController) GenRandomUsername() {
}
