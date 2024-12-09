package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/thats-insane/awt-final/internal/data"
	"github.com/thats-insane/awt-final/internal/validator"
)

/* Add a new review */
func (a *appDependencies) createReviewHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		BookID    int64     `json:"book_id"`
		UserID    int64     `json:"user_id"`
		Rating    int64     `json:"rating"`
		Desc      string    `json:"desc"`
		CreatedAt time.Time `json:"-"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequest(w, r, err)
		return
	}

	review := &data.Review{
		BookID:    incomingData.BookID,
		UserID:    incomingData.UserID,
		Rating:    incomingData.Rating,
		Desc:      incomingData.Desc,
		CreatedAt: incomingData.CreatedAt,
	}
	v := validator.New()
	data.ValidateReview(v, review)
	if !v.IsEmpty() {
		a.failedValidation(w, r, v.Errors)
		return
	}

	err = a.reviewModel.Insert(review)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("api/v1/review/%d", review.ID))
	data := envelope{
		"review": review,
	}

	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	fmt.Fprintf(w, "%+v\n", incomingData)
}

/* Display a review */
func (a *appDependencies) displayReviewHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	review, err := a.reviewModel.Get(id)
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
		"review": review,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
}

/* Update a review */
func (a *appDependencies) updateReviewHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	review, err := a.reviewModel.Get(id)
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
		BookID    *int64     `json:"book_id"`
		UserID    *int64     `json:"user_id"`
		Rating    *int64     `json:"rating"`
		Desc      *string    `json:"desc"`
		CreatedAt *time.Time `json:"-"`
	}
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequest(w, r, err)
		return
	}

	if incomingData.BookID != nil {
		review.BookID = *incomingData.BookID
	}
	if incomingData.UserID != nil {
		review.UserID = *incomingData.UserID
	}
	if incomingData.Rating != nil {
		review.Rating = *incomingData.Rating
	}
	if incomingData.Desc != nil {
		review.Desc = *incomingData.Desc
	}
	if incomingData.CreatedAt != nil {
		review.CreatedAt = *incomingData.CreatedAt
	}

	v := validator.New()
	data.ValidateReview(v, review)
	if !v.IsEmpty() {
		a.failedValidation(w, r, v.Errors)
		return
	}

	err = a.reviewModel.Update(review)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	data := envelope{
		"review": review,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
}

/* Delete a review */
func (a *appDependencies) deleteReviewHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	err = a.reviewModel.Delete(id)
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
		"message": "review successfully deleted",
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErr(w, r, err)
	}
}
