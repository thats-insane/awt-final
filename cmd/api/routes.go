package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *appDependencies) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(a.notFound)
	router.MethodNotAllowed = http.HandlerFunc(a.notAllowed)
	router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", a.healthCheckHandler)

	router.HandlerFunc(http.MethodGet, "/api/v1/books", a.requireActivated(a.listBooksHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/books/:bid", a.requireActivated(a.displayBookHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/books/search", a.requireActivated(a.searchBooksHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/lists", a.requireActivated(a.listListsHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/lists/:id", a.requireActivated(a.displayListHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/books/:bid/reviews", a.requireActivated(a.displayReviewHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/users/:id", a.requireActivated(a.displayUserHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/users/:uid/lists", a.requireActivated(a.displayUserListsHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/users/:id/reviews", a.requireActivated(a.displayUserReviewsHandler))

	router.HandlerFunc(http.MethodPost, "/api/v1/users", a.createUserHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/books", a.requireActivated(a.createBookHandler))
	router.HandlerFunc(http.MethodPost, "/api/v1/lists", a.requireActivated(a.createListHandler))
	router.HandlerFunc(http.MethodPost, "/api/v1/lists/:lid/books", a.requireActivated(a.addBookToListHandler))
	router.HandlerFunc(http.MethodPost, "/api/vi/books/:bid/reviews", a.requireActivated(a.createReviewHandler))
	router.HandlerFunc(http.MethodPost, "/api/v1/tokens/password-reset", a.createPasswordResetTokenHandler)

	router.HandlerFunc(http.MethodPut, "/api/v1/users/activated", a.activateUserHandler)
	router.HandlerFunc(http.MethodPut, "/api/v1/books/:id", a.requireActivated(a.updateBookHandler))
	router.HandlerFunc(http.MethodPut, "/api/v1/lists/:id", a.requireActivated(a.updateListHandler))
	router.HandlerFunc(http.MethodPut, "/api/v1/reviews/:id", a.requireActivated(a.updateReviewHandler))
	router.HandlerFunc(http.MethodPut, "/api/v1/users/password", a.updateUserPasswordHandler)

	router.HandlerFunc(http.MethodDelete, "/api/v1/books/:id", a.requireActivated(a.deleteBookHandler))
	router.HandlerFunc(http.MethodDelete, "/api/v1/lists/:id", a.requireActivated(a.deleteListHandler))
	router.HandlerFunc(http.MethodDelete, "/api/v1/lists/:bid/books", a.requireActivated(a.deleteBookFromListHandler))
	router.HandlerFunc(http.MethodDelete, "/api/v1/reviews/:id", a.requireActivated(a.deleteReviewHandler))

	return a.recoverPanic(a.enableCORS(a.rateLimit(a.authenticate(router))))
}
