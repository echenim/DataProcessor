package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/echenim/data-processor/models"
)

type ScannedDataRepository struct {
	pubSubClient *pubsub.Client
	dbClient     *sql.DB
}

func NewScannedDataRepository(pubsubClient *pubsub.Client, dbClient *sql.DB) *ScannedDataRepository {
	return &ScannedDataRepository{
		pubSubClient: pubsubClient,
		dbClient:     dbClient,
	}
}

// PubSub implmentation
// ProcessBatchScans processes messages in batches rather than individually.
func (r *ScannedDataRepository) ProcessBatchScans(ctx context.Context, batchSize int, batchTimeout time.Duration) {
	subscriptionID := "scan-sub"
	sub := r.pubSubClient.Subscription(subscriptionID)

	// Adjust the overall timeout as needed
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Create a channel to collect messages
	msgChan := make(chan *pubsub.Message)

	go func() {
		err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			// Simply send the message to the channel without processing it here
			msgChan <- msg
		})
		if err != nil {
			log.Fatalf("Receive: %v", err)
		}
	}()

	// Process messages in batches
	for {
		select {
		case <-ctx.Done():
			return // Exit if the context is done
		case msg := <-msgChan:
			batch := make([]*pubsub.Message, 0, batchSize)
			batch = append(batch, msg)

		BatchLoop:
			for i := 1; i < batchSize; i++ {
				select {
				case msg := <-msgChan:
					batch = append(batch, msg)
				case <-time.After(batchTimeout):
					break BatchLoop // Break if no messages are received within the timeout
				}
			}

			// Process the batch
			r.processBatch(ctx, batch)
		}
	}
}

func (r *ScannedDataRepository) processBatch(ctx context.Context, batch []*pubsub.Message) {
	scanresults := make([]models.ScannedResult, 0, len(batch))
	for _, msg := range batch {
		var scan models.Scan
		if err := json.Unmarshal(msg.Data, &scan); err != nil {
			log.Printf("Could not decode message data: %v", err)
			msg.Nack()
			continue
		}

		scanresults = append(scanresults, r.ConvertToScannedResult(scan))
		msg.Ack()
	}

	log.Printf("Processed a batch of %d scans\n", len(scanresults))
	log.Printf("scans records : %v\n", scanresults)
	// Further processing with scans...
	// e.g., bulk insert into a database
}

// ConvertToScannedResult transforms a Scan struct to a ScannedResult struct.
func (r *ScannedDataRepository) ConvertToScannedResult(scan models.Scan) models.ScannedResult {
	return models.ScannedResult{
		IP:        scan.Ip,
		Port:      scan.Port,
		Service:   scan.Service,
		Response:  "Hello world",
		Timestamp: r.unixToTime(scan.Timestamp),
	}
}

// unixToTime converts an int64 Unix timestamp to a time.Time object.
func (*ScannedDataRepository) unixToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// Db implementation

// Explanation:
// Dynamic SQL Construction: The function dynamically constructs a single INSERT SQL statement that can insert multiple rows at once.
// This is done by creating a parameterized placeholder string for each ScannedResult in the batch.
// Parameter Handling: It appends the corresponding values for each placeholder to a slice of interface{}. This slice is then passed to
// ExecContext using variadic syntax (valueArgs...), correctly associating each placeholder in the query with its intended value.
// Batch Upsert: The SQL statement uses the ON CONFLICT clause to handle cases where an inserted row would violate a unique
// constraint on the (ip, port, service) tuple. In such cases, it updates the existing row with the new t                                      imestamp and response.
// Benefits:
// This approach significantly reduces the number of database operations by combining many insert operations into a single query,
// which can improve performance when processing large volumes of data. Additionally, using the ON CONFLICT clause for upserts
// ensures that the database maintains only the latest scan result for each unique combination of IP, port, and service.
func (r *ScannedDataRepository) SaveBatch(ctx context.Context, results []*models.ScannedResult) error {
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

	_, err := r.dbClient.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("error upserting scan results batch: %w", err)
	}
	return nil
}
