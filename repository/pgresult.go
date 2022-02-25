package repository

import (
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type PgResult interface {
	Next() bool
	Scan(dst interface{}) error
	Columns() []string
	Close()
}

type pgResult struct {
	pgx.Rows
}

func newPgResult(rows pgx.Rows) PgResult {
	return &pgResult{rows}
}

func (p *pgResult) Scan(dst interface{}) error {
	rows := pgxscan.NewRowScanner(p.Rows)
	return  rows.Scan(dst)
}

func (p *pgResult) Columns() []string {
	var columns []string
	for _, value := range p.Rows.FieldDescriptions() {
		columns = append(columns, string(value.Name))
	}

	return columns
}