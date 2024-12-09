package main

import (
	"net/http"
	"time"

	"github.com/thats-insane/awt-final/internal/data"
)

func (a *appDependencies) createPasswordResetTokenHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Email string `json:"email"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequest(w, r, err)
		return
	}

	user, err := a.userModel.GetByEmail(incomingData.Email)
	if err != nil {
		a.notFound(w, r)
		return
	}

	token, err := a.tokenModel.New(int64(user.ID), 1*time.Hour, data.ScopeReset)

	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	data := envelope{
		"message": "An email will be sent with password reset instructions",
	}
	a.background(func() {
		emailData := map[string]any{
			"resetToken": token.Plaintext,
			"userID":     user.ID,
		}

		err = a.mailer.Send(user.Email, "reset_password.tmpl", emailData)
		if err != nil {
			a.logger.Error("failed to send password reset email: " + err.Error())
		}
	})

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
}
