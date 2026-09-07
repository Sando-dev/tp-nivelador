package main

import (
	"errors"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	client "github.com/7574-sistemas-distribuidos/tp-nivelador/src/client"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
)

func loadConfig() (client.ClientConfig, error) {
	agencyIdStr := os.Getenv("AGENCY_ID")
	if agencyIdStr == "" {
		return client.ClientConfig{}, errors.New("AGENCY_ID environment variable is required")
	}

	agencyId, err := strconv.Atoi(agencyIdStr)
	if err != nil {
		return client.ClientConfig{}, errors.New("AGENCY_ID must be an integer")
	}

	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		return client.ClientConfig{}, errors.New("SERVER_HOST environment variable is required")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		return client.ClientConfig{}, errors.New("SERVER_PORT environment variable is required")
	}

	inputFile := os.Getenv("INPUT_FILE")
	if inputFile == "" {
		return client.ClientConfig{}, errors.New("INPUT_FILE environment variable is required")
	}

	outputFile := os.Getenv("OUTPUT_FILE")
	if outputFile == "" {
		return client.ClientConfig{}, errors.New("OUTPUT_FILE environment variable is required")
	}

	batchSizeStr := os.Getenv("BATCH_SIZE")
	if batchSizeStr == "" {
		return client.ClientConfig{}, errors.New("BATCH_SIZE environment variable is required")
	}

	batchSize, err := strconv.Atoi(batchSizeStr)
	if err != nil || batchSize <= 0 {
		return client.ClientConfig{}, errors.New(
			"BATCH_SIZE must be a positive integer",
		)
	}

	return client.ClientConfig{
		ServerHost: serverHost,
		ServerPort: serverPort,
		AgencyId:   agencyId,
		InputFile:  inputFile,
		OutputFile: outputFile,
		BatchSize:  batchSize,
	}, nil
}

func run() int {
	shutdown := make(chan struct{})
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		close(shutdown)
	}()

	config, err := loadConfig()
	if err != nil {
		logger.Error("load-config", logger.Fail, "err", err)
		return 1
	}

	client, err := client.NewClient(config, shutdown)
	if err != nil {
		select {
		case <-shutdown:
			return 0
		default:
			logger.Error("client-new", logger.Fail, "err", err)
			return 1
		}
	}

	if err := client.Run(); err != nil {
		select {
		case <-shutdown:
			return 0
		default:
			logger.Error("client-run", logger.Fail, "err", err)
			return 1
		}
	}
	return 0
}

func main() {
	os.Exit(run())
}
