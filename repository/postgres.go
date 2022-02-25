package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	pg "github.com/jackc/pgx/v4/pgxpool"
)

type PostgresQueryer interface {
	Query(ctx context.Context, query string, params ...interface{}) (pgx.Rows, error)
	QueryScan(ctx context.Context, query string, params ...interface{}) (PgResult, error)
	Exec(ctx context.Context, query string, params ...interface{}) (int64, error)
}

type queryer struct {
	conn *pg.Pool
}

func NewPgQueryer(d *pg.Pool) PostgresQueryer {
	return &queryer{
		conn: d,
	}
}

func (q *queryer) Query(ctx context.Context, query string, params ...interface{}) (pgx.Rows, error) {
	res, err := q.conn.Query(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("query: unable to execute query: %s", err)
	}

	return res, nil
}

func (q *queryer) QueryScan(ctx context.Context, query string, params ...interface{}) (PgResult, error) {
	res, err := q.Query(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("queryScan: %s", err)
	}

	return newPgResult(res), nil
}

func (q *queryer) Exec(ctx context.Context, query string, params ...interface{}) (int64, error) {
	res, err := q.conn.Exec(ctx, query, params...)
	if err != nil {
		return -1, fmt.Errorf("exec: unable to execute query: %s", err)
	}

	return res.RowsAffected(), nil
}