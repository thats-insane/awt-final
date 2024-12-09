package data

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/lib/pq"
)

type Perms []string

type PermsModel struct {
	DB *sql.DB
}

func (p Perms) Include(code string) bool {
	return slices.Contains(p, code)
}

func (p PermsModel) GetAll(id int64) (Perms, error) {
	query := `
		SELECT permissions.code
		FROM permissions 
		INNER JOIN users_permissions ON 
		users_permissions.permission_id = permissions.id
		INNER JOIN users ON users_permissions.user_id = users.id
        WHERE users.id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var perms Perms
	for rows.Next() {
		var perm string
		err := rows.Scan(&perm)
		if err != nil {
			return nil, err
		}
		perms = append(perms, perm)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return perms, nil
}

func (p PermsModel) Add(id int64, codes ...string) error {
	query := `
		INSERT INTO users_permissions
        SELECT $1, permissions.id FROM permissions 
        WHERE permissions.code = ANY($2)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := p.DB.ExecContext(ctx, query, id, pq.Array(codes))

	return err
}
