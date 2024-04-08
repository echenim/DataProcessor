package services

import (
	"context"

	"github.com/echenim/data-processor/repositories"
)

type ScannedProcessorService struct {
	processor *repositories.ScannedProcessorRepository
}

func NewScannedProcessorService(_processor *repositories.ScannedProcessorRepository) *ScannedProcessorService {
	return &ScannedProcessorService{processor: _processor}
}

func (s *ScannedProcessorService) ProcessScanData(ctx context.Context) {
	s.processor.ProcessBatchScans(ctx)
}
