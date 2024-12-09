package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/thats-insane/awt-final/internal/validator"
)

type List struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Desc       string `json:"description"`
	UserID     int64  `json:"user_id"`
	BookListID int64  `json:"books"`
	Status     string `json:"status"`
}

type BookList struct {
	ID     int64 `json:"id"`
	ListID int64 `json:"list_id"`
	BookID int64 `json:"book_id"`
}

type ListModel struct {
	DB *sql.DB
}

/* Add a new reading list to the database */
func (l ListModel) Insert(list *List) error {
	query := `
		INSERT INTO lists(name, description, user_id, book_list_id, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	args := []any{list.Name, list.Desc, list.UserID, list.BookListID, list.Status}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return l.DB.QueryRowContext(ctx, query, args...).Scan(&list.ID)

}

/* Select all reading lists from database */
func (l ListModel) GetAll(filters Filters) ([]*List, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT id, name, description, user_id, book_list_id, status
		FROM lists
		ORDER BY %s %s, id ASC
		LIMIT $1 OFFSET $2
	`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := l.DB.QueryContext(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	var totalRecords int
	lists := []*List{}

	for rows.Next() {
		var list List
		err := rows.Scan(&list.ID, &list.Name, &list.Desc, &list.UserID, &list.BookListID, &list.Status)
		if err != nil {
			return nil, Metadata{}, err
		}
		lists = append(lists, &list)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return lists, metadata, nil
}

/* Insert book into list */
func (l ListModel) AddBook(booklist *BookList) error {
	query := `
		INSERT INTO book_list (list_id, book_id)
		VALUES ($1, $2)
		RETURNING id
	`

	args := []any{booklist.ListID, booklist.BookID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return l.DB.QueryRowContext(ctx, query, args...).Scan(&booklist.ID)

}

/* Select specific reading list from database */
func (l ListModel) Get(id int64) (*List, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, description, user_id, book_list_id, status 
		FROM lists
		WHERE id = $1
	`

	var list List
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := l.DB.QueryRowContext(ctx, query, id).Scan(&list.ID, &list.Name, &list.Desc, &list.UserID, &list.BookListID, &list.Status)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &list, nil
}

/* Select all books in the reading list from the database */
func (l ListModel) GetBooks(id int64) (*BookList, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, list_id, book_id
		FROM book_list
		WHERE list_id = $1
	`

	var booklist BookList
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := l.DB.QueryRowContext(ctx, query, id).Scan(&booklist.ID, booklist.ListID, booklist.BookID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &booklist, nil
}

/* Update a reading list's entry */
func (l ListModel) Update(list *List) error {
	query := `
		UPDATE list
		SET name = $1, description = $2, user_id = $3, book_list_id = $4, status = $5
		WHERE id = $6
		RETURNING id
	`

	args := []any{list.Name, list.Desc, &list.UserID, &list.BookListID, &list.Status}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return l.DB.QueryRowContext(ctx, query, args...).Scan(&list.ID)
}

/* Delete a reading list (delete books from the reading list seperately) */
func (l ListModel) Delete(id int64) error {
	if id < 1 {
		return nil
	}

	query := `
		DELETE FROM lists
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := l.DB.ExecContext(ctx, query, id)
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

/* Delete books from reading list */
func (l ListModel) DeleteBook(id int64) error {
	if id < 1 {
		return nil
	}

	query := `
		DELETE FROM book_list
		WHERE book_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := l.DB.ExecContext(ctx, query, id)
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

/* Validation for reading list */
func ValidateList(v *validator.Validator, list *List) {
	v.Check(list.Name != "", "list", "must be provided")
	v.Check(len(list.Name) <= 100, "list", "must not be more than 100 bytes long")
	v.Check(list.UserID > 0, "list", "must be a positive integer")
	v.Check(list.BookListID > 0, "list", "must be a positive integer")
	v.Check(list.Desc != "", "list", "must be provided")
	v.Check(len(list.Desc) <= 225, "list", "must not be more than 225 bytes long")
	v.Check(list.Status == "reading" || list.Status == "finished", "list", "must be reading or finished")
}
