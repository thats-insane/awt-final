package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/thats-insane/awt-final/internal/data"
	"github.com/thats-insane/awt-final/internal/validator"
)

/* Create a new list */
func (a *appDependencies) createListHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Name       string `json:"name"`
		Desc       string `json:"desc"`
		UserID     int64  `json:"user_id"`
		BookListID int64  `json:"books"`
		Status     string `json:"status"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequest(w, r, err)
		return
	}

	list := &data.List{
		Name:       incomingData.Name,
		Desc:       incomingData.Desc,
		UserID:     incomingData.UserID,
		BookListID: incomingData.BookListID,
		Status:     incomingData.Status,
	}

	v := validator.New()
	data.ValidateList(v, list)
	if !v.IsEmpty() {
		a.failedValidation(w, r, v.Errors)
		return
	}

	err = a.listModel.Insert(list)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("api/v1/list/%d", list.ID))
	data := envelope{
		"list": list,
	}

	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	fmt.Fprintf(w, "%+v\n", incomingData)
}

/* Select all lists */
func (a *appDependencies) listListsHandler(w http.ResponseWriter, r *http.Request) {
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

	list, metadata, err := a.listModel.GetAll(queryParametersData.Filters)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	data := envelope{
		"list":      list,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
	}
}

/* Add a new book to reading list */
func (a *appDependencies) addBookToListHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		ListID int64 `json:"list_id"`
		BookID int64 `json:"book_id"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequest(w, r, err)
		return
	}

	booklist := &data.BookList{
		ListID: incomingData.ListID,
		BookID: incomingData.BookID,
	}
	err = a.listModel.AddBook(booklist)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/lists/%d/books", booklist.ID))
	data := envelope{
		"booklist": booklist,
	}

	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	fmt.Fprintf(w, "%+v\n", incomingData)
}

/* Select one reading list */
func (a *appDependencies) displayListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	list, err := a.listModel.Get(id)
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
		"list": list,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
}

/* Update a reading list */
func (a *appDependencies) updateListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	list, err := a.listModel.Get(id)
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
		Name       *string `json:"name"`
		Desc       *string `json:"desc"`
		UserID     *int64  `json:"user_id"`
		BookListID *int64  `json:"books"`
		Status     *string `json:"status"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequest(w, r, err)
		return
	}

	if incomingData.Name != nil {
		list.Name = *incomingData.Name
	}
	if incomingData.Desc != nil {
		list.Desc = *incomingData.Desc
	}
	if incomingData.UserID != nil {
		list.UserID = *incomingData.UserID
	}
	if incomingData.BookListID != nil {
		list.BookListID = *incomingData.BookListID
	}
	if incomingData.Status != nil {
		list.Status = *incomingData.Status
	}

	v := validator.New()
	data.ValidateList(v, list)
	if !v.IsEmpty() {
		a.failedValidation(w, r, v.Errors)
		return
	}

	err = a.listModel.Update(list)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	data := envelope{
		"list": list,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
}

/* Delete a reading list */
func (a *appDependencies) deleteListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	err = a.listModel.Delete(id)
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
		"message": "list successfully deleted",
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
	}
}

/* Delete a book from a reading list */
func (a *appDependencies) deleteBookFromListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	err = a.listModel.DeleteBook(id)
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
		"message": "book successfully deleted from list",
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
	}
}
