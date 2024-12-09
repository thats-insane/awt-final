package main

import (
	"fmt"
	"net/http"
)

func (a *appDependencies) logErr(r *http.Request, err error) {
	method := r.Method
	uri := r.URL.RequestURI()
	a.logger.Error(err.Error(), "method", method, "uri", uri)
}

func (a *appDependencies) errResponseJSON(w http.ResponseWriter, r *http.Request, status int, msg any) {
	errMsg := envelope{
		"error": msg,
	}
	err := a.writeJSON(w, status, errMsg, nil)
	if err != nil {
		a.logErr(r, err)
		w.WriteHeader(500)
	}
}

func (a *appDependencies) serverErr(w http.ResponseWriter, r *http.Request, err error) {
	a.logErr(r, err)
	msg := "the server encountered a problem and could not process your request"
	a.errResponseJSON(w, r, http.StatusInternalServerError, msg)
}

func (a *appDependencies) notFound(w http.ResponseWriter, r *http.Request) {
	msg := "the requested resource could not be found"
	a.errResponseJSON(w, r, http.StatusNotFound, msg)
}

func (a *appDependencies) notAllowed(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	a.errResponseJSON(w, r, http.StatusMethodNotAllowed, msg)
}

func (a *appDependencies) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	a.errResponseJSON(w, r, http.StatusBadRequest, err.Error())
}

func (a *appDependencies) failedValidation(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	a.errResponseJSON(w, r, http.StatusUnprocessableEntity, errors)
}

func (a *appDependencies) rateLimitExceed(w http.ResponseWriter, r *http.Request) {
	msg := "rate limit exceeded"
	a.errResponseJSON(w, r, http.StatusTooManyRequests, msg)
}

func (a *appDependencies) editConflict(w http.ResponseWriter, r *http.Request) {
	msg := "unable to update record to due an edit conflict, try again"
	a.errResponseJSON(w, r, http.StatusConflict, msg)
}

func (a *appDependencies) invalidAuthToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	msg := "invalid/missing authentication token"
	a.errResponseJSON(w, r, http.StatusUnauthorized, msg)
}

func (a *appDependencies) authRequired(w http.ResponseWriter, r *http.Request) {
	msg := "you must be authenticated to access this resource"
	a.errResponseJSON(w, r, http.StatusUnauthorized, msg)
}

func (a *appDependencies) inactiveAccount(w http.ResponseWriter, r *http.Request) {
	msg := "your user account must be activated to access this resource"
	a.errResponseJSON(w, r, http.StatusForbidden, msg)
}

func (a *appDependencies) invalidCredentials(w http.ResponseWriter, r *http.Request) {
	msg := "invalid auth credentials"
	a.errResponseJSON(w, r, http.StatusUnauthorized, msg)
}
