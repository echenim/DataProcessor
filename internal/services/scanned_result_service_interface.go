package services

import "context"

type Service interface {
	ProcessScanData(ctx context.Context)
}
