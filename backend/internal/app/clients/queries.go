package clients

import (
	"context"
	"database/sql"
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

func (q ClientQueries) Insert(ctx context.Context, in CreateClientInput) (int64, error) {
	res, err := q.DB.ExecContext(ctx, `
    INSERT INTO clients (name, company_name, address, email)
    VALUES (?, ?, ?, ?)
  `, in.Name, in.CompanyName, in.Address, in.Email)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}
