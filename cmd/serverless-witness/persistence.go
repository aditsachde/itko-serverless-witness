package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/transparency-dev/witness/omniwitness"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PgPersistence struct {
	ctx    context.Context
	pool   *pgxpool.Pool
	region string
}

func NewPgPersistence(ctx context.Context, url string, region string) (*PgPersistence, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return &PgPersistence{
		ctx,
		pool,
		region,
	}, nil
}

func (p *PgPersistence) Close() {
	p.pool.Close()
}

func (p *PgPersistence) Init() error {
	// check if the table exists
	_, err := p.pool.Exec(p.ctx, `SELECT 1 FROM chkpts LIMIT 1`)
	if err != nil {
		return fmt.Errorf("error checking for chkpts table: %v", err)
	}
	return nil
}

func (p *PgPersistence) Logs() ([]string, error) {
	query := `SELECT log_id FROM chkpts WHERE region = $1`

	rows, err := p.pool.Query(p.ctx, query, p.region)
	if err != nil {
		return nil, fmt.Errorf("error querying log ids: %v", err)
	}
	defer rows.Close()

	var logIds []string
	for rows.Next() {
		var logID string
		if err := rows.Scan(&logID); err != nil {
			return nil, fmt.Errorf("error scanning log id: %v", err)
		}
		logIds = append(logIds, logID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return logIds, nil
}

func (p *PgPersistence) ReadOps(logId string) (omniwitness.LogStateReadOps, error) {
	return &PgLogStateReadOps{
		logId,
		p,
	}, nil
}

func (p *PgPersistence) WriteOps(logId string) (omniwitness.LogStateWriteOps, error) {
	tx, err := p.pool.Begin(p.ctx)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	return &PgLogStateWriteOps{
		logId,
		tx,
		p,
	}, nil
}

type PgLogStateReadOps struct {
	logId string
	p     *PgPersistence
}

func (r *PgLogStateReadOps) GetLatest() ([]byte, []byte, error) {
	return getLatestCheckpoint(r.p.pool.QueryRow, r.p.ctx, r.logId, r.p.region)
}

type PgLogStateWriteOps struct {
	logId string
	tx    pgx.Tx
	p     *PgPersistence
}

func (w *PgLogStateWriteOps) GetLatest() ([]byte, []byte, error) {
	return getLatestCheckpoint(w.tx.QueryRow, w.p.ctx, w.logId, w.p.region)
}

func (w *PgLogStateWriteOps) Set(chkpt []byte, rng []byte) error {
	_, err := w.tx.Exec(w.p.ctx,
		`INSERT INTO chkpts (region, log_id, chkpt, range) VALUES ($1, $2, $3, $4) 
		 ON CONFLICT (region, log_id) DO UPDATE SET chkpt = $3, range = $4`,
		w.p.region, w.logId, chkpt, rng)

	if err != nil {
		return fmt.Errorf("error setting checkpoint: %v", err)
	}
	return w.tx.Commit(w.p.ctx)
}

func (w *PgLogStateWriteOps) Close() error {
	return w.tx.Rollback(w.p.ctx)
}

func getLatestCheckpoint(
	queryRow func(ctx context.Context, sql string, args ...any) pgx.Row,
	ctx context.Context,
	logId string,
	region string,
) ([]byte, []byte, error) {
	row := queryRow(ctx, "SELECT chkpt, range FROM chkpts WHERE region = $1 AND log_id = $2", region, logId)

	var chkpt []byte
	var rnge []byte
	err := row.Scan(&chkpt, &rnge)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, status.Errorf(codes.NotFound, "no checkpoint for log %q", logId)
		}
		return nil, nil, err
	}
	return chkpt, rnge, nil
}
