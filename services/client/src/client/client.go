package client

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/bet"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const CONNECTION_ATTEMPTS_MAX = 5
const CONNECTION_ATTEMPS_DELAY_MS = 500

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   int
	InputFile  string
	OutputFile string
	BatchSize  int
}

type Client struct {
	conn     net.Conn
	config   ClientConfig
	shutdown <-chan struct{}
}

func NewClient(
	config ClientConfig,
	shutdown <-chan struct{},
) (*Client, error) {

	conn, err := connectToServer(
		config.ServerHost,
		config.ServerPort,
		shutdown,
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:     conn,
		config:   config,
		shutdown: shutdown,
	}, nil
}

func connectToServer(
	host string,
	port string,
	shutdown <-chan struct{},
) (net.Conn, error) {

	const action = "connect-to-server"
	logger.Info(action, logger.InProgress)

	var lastErr error

	for i := 0; i < CONNECTION_ATTEMPTS_MAX; i++ {

		select {
		case <-shutdown:
			return nil, nil
		default:
		}

		conn, err := net.Dial("tcp", host+":"+port)
		if err == nil {
			logger.Info(action, logger.Success)
			return conn, nil
		}

		lastErr = err

		logger.Warn(
			action,
			logger.Fail,
			"attempt",
			i,
		)

		select {
		case <-shutdown:
			return nil, nil

		case <-time.After(
			CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond,
		):
		}
	}

	return nil, lastErr
}

func (client *Client) Run() error {
	const mainAction = "test-echo-server"

	if client.conn == nil {
		return nil
	}

	// Si llega SIGTERM mientras estamos bloqueados en Read/Write,
	// cerramos el socket para desbloquear la operación.
	go func() {
		<-client.shutdown
		client.conn.Close()
	}()

	defer client.conn.Close()

	inputFile, err := os.Open(client.config.InputFile)
	if err != nil {
		if client.isShuttingDown() {
			return nil
		}
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(client.config.OutputFile)
	if err != nil {
		if client.isShuttingDown() {
			return nil
		}
		return err
	}
	defer outputFile.Close()

	reader := csv.NewReader(inputFile)
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	if err := client.sendBetsInBatches(
		reader,
		client.config.BatchSize,
	); err != nil {
		if client.isShuttingDown() {
			return nil
		}
		return err
	}

	if client.isShuttingDown() {
		return nil
	}

	if err := safe_socket.SendAll(
		client.conn,
		protocol.SerializeFinish(),
	); err != nil {
		if client.isShuttingDown() {
			return nil
		}
		return err
	}

	if err := client.receiveWinners(writer); err != nil {
		if client.isShuttingDown() {
			return nil
		}
		return err
	}

	logger.Info(
		mainAction,
		logger.Success,
		"agency-id",
		client.config.AgencyId,
	)

	return nil
}

func (client *Client) isShuttingDown() bool {
	select {
	case <-client.shutdown:
		return true
	default:
		return false
	}
}

func (client *Client) sendBetsInBatches(
	reader *csv.Reader,
	batchSize int,
) error {

	batch := make([]*bet.Bet, 0, batchSize)

	for {
		if client.isShuttingDown() {
			return nil
		}

		row, err := reader.Read()

		if err == io.EOF {
			if len(batch) > 0 {
				if err := client.sendBatch(batch); err != nil {
					return err
				}
			}
			return nil
		}

		if err != nil {
			return err
		}

		newBet, err := client.betFromRow(row)
		if err != nil {
			return err
		}

		batch = append(batch, newBet)

		if len(batch) == batchSize {
			if err := client.sendBatch(batch); err != nil {
				return err
			}

			batch = batch[:0]
		}
	}
}

func (client *Client) sendBatch(batch []*bet.Bet) error {
	packet := protocol.SerializeBatch(batch)

	if err := safe_socket.SendAll(client.conn, packet); err != nil {
		return err
	}

	messageType, _, err := protocol.ReceiveMessage(client.conn)
	if err != nil {
		return err
	}

	if messageType == protocol.MessageBatch_Ok {
		return nil
	}

	if messageType == protocol.MessageBatch_Error {
		return fmt.Errorf("server failed to process batch")
	}

	return fmt.Errorf("unexpected response type: %d", messageType)
}

func (client *Client) betFromRow(
	row []string,
) (*bet.Bet, error) {

	documentation, err := strconv.Atoi(row[2])
	if err != nil {
		return nil, err
	}

	number, err := strconv.Atoi(row[4])
	if err != nil {
		return nil, err
	}

	return bet.NewBet(
		client.config.AgencyId,
		row[0],
		row[1],
		documentation,
		row[3],
		number,
	), nil
}

func (client *Client) receiveWinners(
	writer *csv.Writer,
) error {

	for {
		if client.isShuttingDown() {
			return nil
		}

		messageType, payload, err :=
			protocol.ReceiveMessage(client.conn)

		if err != nil {
			return err
		}

		if messageType == protocol.MessageFinish {
			return nil
		}

		if messageType != protocol.MessageWinners {
			continue
		}

		winner, err := protocol.DeserializeBet(payload)
		if err != nil {
			return err
		}

		if err := writer.Write([]string{
			winner.Name,
			winner.LastName,
			strconv.Itoa(winner.Document),
			winner.Birthday,
			strconv.Itoa(winner.Number),
		}); err != nil {
			return err
		}
	}
}
