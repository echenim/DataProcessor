package backup

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/echenim/data-processor/dal"
	"github.com/echenim/data-processor/logger"
	"github.com/echenim/data-processor/models"
)

type Processor struct {
	dal *dal.DAL
}

// func NewProcessor(ctx context.Context, cfg *config.Config) (*Processor, error) {
// 	d, err := dal.NewDAL(cfg.DBConnectionString)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to initialize DAL: %w", err)
// 	}

// 	return &Processor{dal: d}, nil
// }

func (p *Processor) ProcessMessageSingle(ctx context.Context, msg *pubsub.Message) {
	defer msg.Ack()

	var messageData struct {
		DataVersion int             `json:"data_version"`
		Data        json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(msg.Data, &messageData); err != nil {
		logger.Error("Failed to unmarshal message:", err)
		return
	}

	var response string
	switch messageData.DataVersion {
	case 1:
		var data struct {
			ResponseBytesBase64 string `json:"response_bytes_utf8"`
		}
		if err := json.Unmarshal(messageData.Data, &data); err != nil {
			logger.Error("Failed to unmarshal data for version 1:", err)
			return
		}
		decodedBytes, err := base64.StdEncoding.DecodeString(data.ResponseBytesBase64)
		if err != nil {
			logger.Error("Failed to decode base64:", err)
			return
		}
		response = string(decodedBytes)
	case 2:
		var data struct {
			ResponseStr string `json:"response_str"`
		}
		if err := json.Unmarshal(messageData.Data, &data); err != nil {
			logger.Error("Failed to unmarshal data for version 2:", err)
			return
		}
		response = data.ResponseStr
	default:
		logger.Error("Unknown data version:", messageData.DataVersion)
		return
	}

	// Example IP, Port, Service values. Replace with actual values from the message.
	scanResult := models.ScannedResult{
		IP:        "192.168.1.1",
		Port:      80,
		Service:   "HTTP",
		Timestamp: time.Now(),
		Response:  response,
	}

	if err := p.dal.UpsertScanResult(ctx, scanResult); err != nil {
		logger.Error("Failed to upsert scan result:", err)
	}
}
