package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"

	"github.com/echenim/data-processor/models"
)

type ScanResultRepository struct {
	db *sql.DB
}

func NewScanResultRepository(db *sql.DB) *ScanResultRepository {
	return &ScanResultRepository{db: db}
}

func (r *ScanResultRepository) Save(ctx context.Context, result *models.ScannedResult) error {
	// SQL query template for inserting a new scan result or updating an existing one based on a conflict
	query := `
    INSERT INTO scan_results (ip, port, service, timestamp, response)
    VALUES ($1, $2, $3, $4, $5)
    ON CONFLICT (ip, port, service) DO UPDATE 
    SET timestamp = excluded.timestamp, 
        response = excluded.response;
    `
	// Execute the query with the context and scan result parameters
	_, err := r.db.ExecContext(ctx, query, result.IP, result.Port, result.Service, result.Timestamp, result.Response)
	// Return an error formatted to include the underlying error if the execution fails
	if err != nil {
		return fmt.Errorf("error upserting scan result: %w", err)
	}
	return nil
}

// Explanation:
// Dynamic SQL Construction: The function dynamically constructs a single INSERT SQL statement that can insert multiple rows at once.
// This is done by creating a parameterized placeholder string for each ScannedResult in the batch.
// Parameter Handling: It appends the corresponding values for each placeholder to a slice of interface{}. This slice is then passed to
// ExecContext using variadic syntax (valueArgs...), correctly associating each placeholder in the query with its intended value.
// Batch Upsert: The SQL statement uses the ON CONFLICT clause to handle cases where an inserted row would violate a unique
// constraint on the (ip, port, service) tuple. In such cases, it updates the existing row with the new timestamp and response.

// Benefits:
// This approach significantly reduces the number of database operations by combining many insert operations into a single query,
// which can improve performance when processing large volumes of data. Additionally, using the ON CONFLICT clause for upserts
// ensures that the database maintains only the latest scan result for each unique combination of IP, port, and service.
func (r *ScanResultRepository) SaveBatch(ctx context.Context, results []*models.ScannedResult) error {
	if len(results) == 0 {
		return nil // No results to process
	}

	// Construct the VALUES list for the INSERT statement dynamically
	valueStrings := make([]string, 0, len(results))
	valueArgs := make([]interface{}, 0, len(results)*5)
	for i, result := range results {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		valueArgs = append(valueArgs, result.IP)
		valueArgs = append(valueArgs, result.Port)
		valueArgs = append(valueArgs, result.Service)
		valueArgs = append(valueArgs, result.Timestamp)
		valueArgs = append(valueArgs, result.Response)
	}

	query := fmt.Sprintf(`
	INSERT INTO scan_results (ip, port, service, timestamp, response)
	VALUES %s
	ON CONFLICT (ip, port, service) DO UPDATE 
	SET timestamp = excluded.timestamp, 
	    response = excluded.response;
	`, strings.Join(valueStrings, ","))

	_, err := r.db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("error upserting scan results batch: %w", err)
	}
	return nil
}
