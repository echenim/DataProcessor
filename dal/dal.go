package dal

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/echenim/data-processor/models"
)

type DAL struct {
	db *sql.DB
}

func NewDAL(driver, conn string) (*DAL, error) {
	db, err := sql.Open(driver, conn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	return &DAL{db: db}, nil
}

func (d *DAL) UpsertScanResult(ctx context.Context, result models.ScannedResult) error {
	query := `
	INSERT INTO scan_results (ip, port, service, timestamp, response)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (ip, port, service) DO UPDATE 
	SET timestamp = excluded.timestamp, 
	    response = excluded.response;
	`
	_, err := d.db.ExecContext(ctx, query, result.IP, result.Port, result.Service, result.Timestamp, result.Response)
	if err != nil {
		return fmt.Errorf("error upserting scan result: %w", err)
	}
	return nil
}
