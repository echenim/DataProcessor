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
	"github.com/echenim/data-processor/internal/models"
)

type ScannedProcessorRepository struct {
	pubSubClient *pubsub.Client
	dbClient     *sql.DB
}

func NewScannedProcessorRepository(pubsubClient *pubsub.Client, dbClient *sql.DB) *ScannedProcessorRepository {
	return &ScannedProcessorRepository{
		pubSubClient: pubsubClient,
		dbClient:     dbClient,
	}
}

// ProcessBatchScans listens for messages from a Pub/Sub subscription and processes them in batches.
func (r *ScannedProcessorRepository) ProcessBatchScans(ctx context.Context, subscriptionID string, batchSize int, batchTimeout time.Duration) {
	// Define the subscription ID, batch size, and batch processing timeout.
	// subscriptionID := "scan-sub"
	// batchSize := 100
	// batchTimeout := 2 * time.Minute

	// Get the subscription from the Pub/Sub client.
	sub := r.pubSubClient.Subscription(subscriptionID)

	// Set a longer timeout for the entire batch processing operation.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel() // Ensure any resources are freed on exit.

	// Create a channel to receive messages from the subscription.
	msgChan := make(chan *pubsub.Message)

	// Start a goroutine to receive messages and send them to the channel.
	go func() {
		err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			// Forward messages to the channel without processing them here.
			msgChan <- msg
		})
		if err != nil {
			log.Fatalf("Receive: %v", err)
		}
	}()

	// Continuously process messages until the context is done.
	for {
		select {
		case <-ctx.Done():
			return // Terminate if the context's deadline is exceeded or cancelled.
		case msg := <-msgChan:
			// Initialize the batch with the first received message.
			batch := make([]*pubsub.Message, 0, batchSize)
			batch = append(batch, msg)

			// Attempt to fill the batch until the batch size is reached or a timeout occurs.
		BatchLoop:
			for i := 1; i < batchSize; i++ {
				select {
				case msg := <-msgChan:
					batch = append(batch, msg)
				case <-time.After(batchTimeout):
					break BatchLoop // Exit the loop if the timeout is reached before the batch is full.
				}
			}

			// Process the messages in the batch.
			go r.processBatch(ctx, batch)
		}
	}
}

// processBatch processes a batch of Pub/Sub messages by converting them into scanned results and inserting them into a database.
func (r *ScannedProcessorRepository) processBatch(ctx context.Context, batch []*pubsub.Message) {
	processedResult := make([]models.ScannedResult, 0, len(batch))
	for _, msg := range batch {
		// Decode each message into a Scan object.
		var scan models.Scan
		if err := json.Unmarshal(msg.Data, &scan); err != nil {
			log.Printf("Could not decode message data: %v", err)
			msg.Nack() // Signal that the message was not processed successfully.
			continue
		}

		// Convert the decoded message to a ScannedResult and add it to the batch.
		processedResult = append(processedResult, r.convertToScannedResult(scan))
		msg.Ack() // Acknowledge successful processing of the message.
	}

	// Insert the processed results into the database as a batch.
	r.insertBatch(ctx, processedResult)
}

// ConvertToScannedResult transforms a Scan struct to a ScannedResult struct.
func (r *ScannedProcessorRepository) convertToScannedResult(scan models.Scan) models.ScannedResult {
	return models.ScannedResult{
		IP:        scan.Ip,
		Port:      scan.Port,
		Service:   scan.Service,
		Response:  "Hello world",
		Timestamp: r.unixToTime(scan.Timestamp),
	}
}

// unixToTime converts an int64 Unix timestamp to a time.Time object.
func (*ScannedProcessorRepository) unixToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

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

// insertBatch inserts a batch of scanned results into the database.
func (r *ScannedProcessorRepository) insertBatch(ctx context.Context, results []models.ScannedResult) {
	// Check if the results slice is empty and log a message if it is.
	if len(results) == 0 {
		log.Print("\nNo results to process\n")
		return // Return early if there are no results to process.
	}

	// Prepare the placeholders and arguments for the batch insert query.
	// Dynamically construct the list of placeholders for each result.
	valueStrings := make([]string, 0, len(results))
	valueArgs := make([]interface{}, 0, len(results)*5)
	for i, result := range results {
		// For each result, add a placeholder string in the format "($1, $2, $3, $4, $5)"
		// where the numbers will be replaced by actual values for each column.
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		// Append the actual values for each placeholder to the valueArgs slice.
		// These will replace the placeholders in the query.
		valueArgs = append(valueArgs, result.IP, result.Port, result.Service, result.Timestamp, result.Response)
	}

	// Construct the final SQL query string with placeholders for the batch insert.
	// Use strings.Join to concatenate the individual value strings with commas.
	query := fmt.Sprintf(`
	INSERT INTO scan_results (ip, port, service, timestamp, response)
	VALUES %s
	ON CONFLICT (ip, port, service) DO UPDATE 
    SET timestamp = excluded.timestamp, 
       response = excluded.response;
	`, strings.Join(valueStrings, ","))

	// Execute the batch insert query against the database.
	rs, err := r.dbClient.ExecContext(ctx, query, valueArgs...)
	// Check if there was an error during the execution of the query.
	if err != nil {
		// Log the error if the query execution failed.
		log.Printf("\nerror upserting scan results batch: %v\n", err)
		return // Return early in case of an error.
	}

	// Log a success message along with the result of the query execution.
	log.Printf("\nscan results batch successful: %v\n", rs)
}
