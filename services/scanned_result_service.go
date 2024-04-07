package services

import (
	"context"

	"github.com/echenim/data-processor/models"
	"github.com/echenim/data-processor/repositories"
)

type ScanResultService struct {
	repo *repositories.ScanResultRepository
}

func NewScanResultService(repo *repositories.ScanResultRepository) *ScanResultService {
	return &ScanResultService{repo: repo}
}

func (s *ScanResultService) SaveScanResult(ctx context.Context, result *models.ScannedResult) error {
	// Implement the logic to call repo.Save
	return s.repo.Save(ctx, result)
}
