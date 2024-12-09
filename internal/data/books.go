package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/thats-insane/awt-final/internal/validator"
)

type Book struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	ISBN      string    `json:"isbn"`
	PubDate   time.Time `json:"pub_date"`
	Genre     string    `json:"genre"`
	Desc      string    `json:"description"`
	AvgRating float64   `json:"avg_rating"`
}

type BookModel struct {
	DB *sql.DB
}

/* Add a new book */
func (b BookModel) Insert(book *Book) error {
	query := `
		INSERT INTO books (title, author, isbn, publication_date, genre, description, average_rating) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	args := []any{book.Title, book.Author, book.ISBN, book.PubDate, book.Genre, book.Desc, book.AvgRating}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return b.DB.QueryRowContext(ctx, query, args...).Scan(&book.ID)
}

/* Select a book */
func (b BookModel) Get(id int64) (*Book, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, title, author, isbn, publication_date, genre, description, average_rating
		FROM books
		WHERE id = $1
	`

	var book Book
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := b.DB.QueryRowContext(ctx, query, id).Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Author, &book.PubDate, &book.Genre, &book.Desc, &book.AvgRating)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &book, nil
}

/* Select all books */
func (b BookModel) GetAll(filters Filters) ([]*Book, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT id, title, author, isbn, publication_date, genre, description, average_rating
		FROM books
		ORDER BY %s %s, id ASC
		LIMIT $1 OFFSET $2
		`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := b.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var totalRecords int
	books := []*Book{}

	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.PubDate, &book.Genre, &book.Desc, &book.AvgRating)
		if err != nil {
			return nil, Metadata{}, err
		}
		books = append(books, &book)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return books, metadata, nil
}

/* Update a book */
func (b BookModel) Update(book *Book, id int64) error {
	query := `
		UPDATE books
		SET title = $1, author = $2, isbn = $3, publication_date = $3, genre = $4, description = $5, average_rating = $6
		WHERE id = $7
		RETURNING id
	`

	args := []any{book.Title, book.Author, book.ISBN, book.PubDate, book.Genre, book.Desc, book.AvgRating, book.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return b.DB.QueryRowContext(ctx, query, args...).Scan(&book.ID)
}

/* Delete a book */
func (b BookModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM books
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := b.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

/* Select a book using filters */
func (b BookModel) Search(title string, author string, genre string, filters Filters) ([]*Book, Metadata, error) {
	query := `
        SELECT id, title, author, isbn, publication_date, genre, description, average_rating
		FROM books
        WHERE (to_tsvector('simple', title) @@
              plainto_tsquery('simple', $1) OR $1 = '')
        AND (to_tsvector('simple', author) @@
             plainto_tsquery('simple', $2) OR $2 = '')
		AND (to_tsvector('simple', genre) @@
             plainto_tsquery('simple', $3) OR $3 = '')
        ORDER BY id
		LIMIT $4 OFFSET $5
     `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := b.DB.QueryContext(ctx, query, title, author, genre)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var totalRecords int
	books := []*Book{}

	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.PubDate, &book.Genre, &book.Desc, &book.AvgRating)
		if err != nil {
			return nil, Metadata{}, err
		}
		books = append(books, &book)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return books, metadata, nil
}

/* Validation for book */
func ValidateBook(v *validator.Validator, book *Book) {
	v.Check(book.Title != "", "book", "must be provided")
	v.Check(len(book.Title) <= 100, "book", "must not be more than 100 bytes long")
	v.Check(book.ISBN != "", "book", "must be provided")
	v.Check(book.Genre != "", "book", "must be provided")
	v.Check(book.Desc != "", "book", "must be provided")
	v.Check(len(book.Desc) <= 225, "book", "must not be more than 225 bytes long")
	v.Check(book.AvgRating >= 1 && book.AvgRating <= 5, "book", "must be between 1 and 5")
}
