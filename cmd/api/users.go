package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/thats-insane/awt-final/internal/data"
	"github.com/thats-insane/awt-final/internal/validator"
)

/* Register a new user, create their token and send them an activation email */
func (a *appDependencies) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := a.readJSON(w, r, incomingData)
	if err != nil {
		a.badRequest(w, r, err)
		return
	}

	user := &data.User{
		Username:  incomingData.Username,
		Email:     incomingData.Email,
		Activated: false,
	}
	err = user.Password.Set(incomingData.Password)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateUser(v, user)
	if !v.IsEmpty() {
		a.failedValidation(w, r, v.Errors)
		return
	}

	err = a.userModel.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists")
			a.failedValidation(w, r, v.Errors)
		default:
			a.serverErr(w, r, err)
		}
		return
	}

	token, err := a.tokenModel.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	data := envelope{
		"user": user,
	}

	a.background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}
		err = a.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			a.logger.Error(err.Error())
		}
	})

	err = a.writeJSON(w, http.StatusCreated, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
}

/* Activate user using their auth token */
func (a *appDependencies) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Plaintext string `json:"token"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequest(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateTokenPlaintext(v, incomingData.Plaintext)
	if !v.IsEmpty() {
		a.failedValidation(w, r, v.Errors)
		return
	}

	user, err := a.userModel.GetForToken(data.ScopeActivation, incomingData.Plaintext)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid/expired activation token")
			a.failedValidation(w, r, v.Errors)
		default:
			a.serverErr(w, r, err)
		}
		return
	}

	user.Activated = true
	err = a.userModel.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			a.editConflict(w, r)
		default:
			a.serverErr(w, r, err)
		}
		return
	}

	data := envelope{
		"user": user,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
	}
}

/* Grab user from database and display */
func (a *appDependencies) displayUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	user, err := a.userModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFound(w, r)
		default:
			a.serverErr(w, r, err)
		}
		return
	}

	data := envelope{
		"user": user,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
}

/* Display users and lists they have made */
func (a *appDependencies) displayUserListsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	booklist, err := a.listModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFound(w, r)
		default:
			a.serverErr(w, r, err)
		}
		return
	}

	books, err := a.listModel.GetBooks(booklist.BookListID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFound(w, r)
		default:
			a.serverErr(w, r, err)
		}
		return
	}

	data := envelope{
		"lists": books,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
}

/* Display users and reviews they have made */
func (a *appDependencies) displayUserReviewsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	userreviews, err := a.reviewModel.GetUser(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFound(w, r)
		default:
			a.serverErr(w, r, err)
		}
		return
	}

	// iterate over all the user reviews and match them to the book
	var data map[string]any
	for i, review := range userreviews {
		reviews := envelope{
			"bookID": review.BookID,
		}
		data[fmt.Sprintf("review%d", i)] = reviews
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
}

func (a *appDependencies) updateUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Password string `json:"password"`
		Token    string `json:"token"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequest(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateTokenPlaintext(v, incomingData.Token)
	if !v.IsEmpty() {
		a.failedValidation(w, r, v.Errors)
		return
	}
	data.ValidatePasswordPlaintext(v, incomingData.Password)
	if !v.IsEmpty() {
		a.failedValidation(w, r, v.Errors)
		return
	}

	user, err := a.userModel.GetForToken(data.ScopeReset, incomingData.Token)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid/expired token")
			a.failedValidation(w, r, v.Errors)
		default:
			a.serverErr(w, r, err)
		}
		return
	}

	user, err = a.userModel.Get(user.ID)
	if err != nil {
		a.notFound(w, r)
		return
	}

	err = user.Password.Set(incomingData.Password)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	data.ValidateUser(v, user)
	if !v.IsEmpty() {
		a.failedValidation(w, r, v.Errors)
		return
	}

	err = a.userModel.Update(user)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	err = a.tokenModel.DeleteAllForUser(data.ScopeReset, user.ID)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	data := envelope{
		"message": "password successfully reset",
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
}
