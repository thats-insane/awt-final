package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/thats-insane/awt-final/internal/data"
	"github.com/thats-insane/awt-final/internal/validator"
)

/* Create a new book */
func (a *appDependencies) createBookHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Title     string    `json:"title"`
		Author    string    `json:"author"`
		ISBN      string    `json:"isbn"`
		PubDate   time.Time `json:"pub_date"`
		Genre     string    `json:"genre"`
		Desc      string    `json:"desc"`
		AvgRating float64   `json:"avg_rating"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequest(w, r, err)
		return
	}

	book := &data.Book{
		Title:     incomingData.Title,
		Author:    incomingData.Author,
		ISBN:      incomingData.ISBN,
		Genre:     incomingData.Genre,
		Desc:      incomingData.Desc,
		PubDate:   incomingData.PubDate,
		AvgRating: incomingData.AvgRating,
	}
	v := validator.New()
	data.ValidateBook(v, book)
	if !v.IsEmpty() {
		a.failedValidation(w, r, v.Errors)
		return
	}

	err = a.bookModel.Insert(book)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("api/v1/book/%d", book.ID))
	data := envelope{
		"book": book,
	}

	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	fmt.Fprintf(w, "%+v\n", incomingData)
}

/* Display a book */
func (a *appDependencies) displayBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	book, err := a.bookModel.Get(id)
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
		"book": book,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
}

/* Update a book */
func (a *appDependencies) updateBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	book, err := a.bookModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFound(w, r)
		default:
			a.serverErr(w, r, err)
		}
		return
	}

	var incomingData struct {
		Title     *string    `json:"title"`
		Author    *string    `json:"author"`
		PubDate   *time.Time `json:"pub_date"`
		ISBN      *string    `json:"isbn"`
		Genre     *string    `json:"genre"`
		Desc      *string    `json:"desc"`
		AvgRating *int       `json:"avg_rating"`
	}
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequest(w, r, err)
		return
	}

	if incomingData.Title != nil {
		book.Title = *incomingData.Title
	}
	if incomingData.Author != nil {
		book.Author = *incomingData.Author
	}
	if incomingData.ISBN != nil {
		book.ISBN = *incomingData.ISBN
	}
	if incomingData.PubDate != nil {
		book.PubDate = *incomingData.PubDate
	}
	if incomingData.Genre != nil {
		book.Genre = *incomingData.Genre
	}
	if incomingData.Desc != nil {
		book.Desc = *incomingData.Desc
	}

	v := validator.New()
	data.ValidateBook(v, book)
	if !v.IsEmpty() {
		a.failedValidation(w, r, v.Errors)
		return
	}

	err = a.bookModel.Update(book, id)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	data := envelope{
		"book": book,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
}

/* Delete a book */
func (a *appDependencies) deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	err = a.bookModel.Delete(id)
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
		"message": "book successfully deleted",
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
	}
}

/* List all books */
func (a *appDependencies) listBooksHandler(w http.ResponseWriter, r *http.Request) {
	var queryParametersData struct {
		data.Filters
	}
	queryParameters := r.URL.Query()
	queryParametersData.Filters.Sort = a.getSingleQueryParameters(queryParameters, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "-id"}
	v := validator.New()
	queryParametersData.Filters.Page = a.getSingleIntegerParameters(queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameters(queryParameters, "page_size", 10, v)
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidation(w, r, v.Errors)
		return
	}

	book, metadata, err := a.bookModel.GetAll(queryParametersData.Filters)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	data := envelope{
		"book":      book,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
	}
}

/* Search for a book based on filters */
func (a *appDependencies) searchBooksHandler(w http.ResponseWriter, r *http.Request) {
	var queryParametersData struct {
		Title  string
		Author string
		Genre  string
		data.Filters
	}
	queryParameters := r.URL.Query()
	queryParametersData.Title = a.getSingleQueryParameters(queryParameters, "title", "")
	queryParametersData.Author = a.getSingleQueryParameters(queryParameters, "author", "")
	queryParametersData.Genre = a.getSingleQueryParameters(queryParameters, "genre", "")
	queryParametersData.Filters.Sort = a.getSingleQueryParameters(queryParameters, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "title", "author", "genre", "-id", "-title", "-author", "-genre"}
	v := validator.New()
	queryParametersData.Filters.Page = a.getSingleIntegerParameters(queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameters(queryParameters, "page_size", 10, v)
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidation(w, r, v.Errors)
		return
	}

	book, metadata, err := a.bookModel.Search(queryParametersData.Title, queryParametersData.Author, queryParametersData.Genre, queryParametersData.Filters)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	data := envelope{
		"book":      book,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
	}
}
