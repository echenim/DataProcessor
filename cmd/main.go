package main

import (
	"context"

	"github.com/echenim/data-processor/clients"
	"github.com/echenim/data-processor/config"
	"github.com/echenim/data-processor/logger"
	"github.com/echenim/data-processor/processor"
)

func main() {
	logger.Setup()
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration:", err)
		return
	}

	ctx := context.Background()
	processor, err := processor.NewProcessor(ctx, cfg)
	if err != nil {
		logger.Error("Failed to initialize processor:", err)
		return
	}

	clients.StartSubscriber(ctx, cfg, processor.ProcessMessage)
}
