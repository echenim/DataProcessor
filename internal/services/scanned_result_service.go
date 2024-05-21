package services

import (
	"context"
	"time"

	"github.com/echenim/data-processor/internal/engine"
)

type ScannedProcessorService struct {
	processor *repositories.ScannedProcessorRepository
}

func NewScannedProcessorService(_processor *repositories.ScannedProcessorRepository) *ScannedProcessorService {
	return &ScannedProcessorService{processor: _processor}
}

func (s *ScannedProcessorService) ProcessScanData(ctx context.Context, subscriptionID string, batchSize int, batchTimeout time.Duration) {
	s.processor.ProcessBatchScans(ctx, subscriptionID, batchSize, batchTimeout)
}
