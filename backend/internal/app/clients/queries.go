package clients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type ClientQueries struct {
	DB *sql.DB
}

func (q ClientQueries) ListClients(ctx context.Context) ([]Client, error) {
	rows, err := q.DB.QueryContext(ctx, `
		SELECT id, name, company_name, address, email, created_at, updated_at
		FROM clients
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Client
	for rows.Next() {
		var c Client
		if err := rows.Scan(
			&c.ID, &c.Name, &c.CompanyName, &c.Address, &c.Email, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (q ClientQueries) Insert(ctx context.Context, in ClientInput) (int64, error) {
	res, err := q.DB.ExecContext(ctx, `
    INSERT INTO clients (name, company_name, address, email)
    VALUES (?, ?, ?, ?)
  `, in.Name, in.CompanyName, in.Address, in.Email)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (q ClientQueries) Delete(ctx context.Context, id int64) (int64, error) {
	res, err := q.DB.ExecContext(ctx, `
		DELETE FROM clients 
		WHERE id = ?
	`, id)

	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

// Fetch one client so handler can return the updated record
func (q ClientQueries) GetByID(ctx context.Context, id int64) (Client, error) {
	var c Client
	err := q.DB.QueryRowContext(ctx, `
		SELECT id, name, company_name, address, email, created_at, updated_at
		FROM clients
		WHERE id = ?
	`, id).Scan(&c.ID, &c.Name, &c.CompanyName, &c.Address, &c.Email, &c.CreatedAt, &c.UpdatedAt)

	return c, err
}

// Partial update
func (q ClientQueries) Update(ctx context.Context, id int64, in UpdateClientInput) (int64, error) {
	setParts := make([]string, 0, 5)
	args := make([]any, 0, 6)

	if in.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *in.Name)
	}
	if in.CompanyName != nil {
		setParts = append(setParts, "company_name = ?")
		args = append(args, *in.CompanyName)
	}
	if in.Address != nil {
		setParts = append(setParts, "address = ?")
		args = append(args, *in.Address)
	}
	if in.Email != nil {
		setParts = append(setParts, "email = ?")
		args = append(args, *in.Email)
	}

	// nothing to update
	if len(setParts) == 0 {
		return 0, errors.New("no fields to update")
	}

	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE clients
		SET %s
		WHERE id = ?
	`, strings.Join(setParts, ", "))

	res, err := q.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
