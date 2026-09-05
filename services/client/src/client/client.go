package client

import (
	"net"
	"time"
	"os"
	"io"
	"encoding/csv"
	"strconv"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/bet"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200

const ECHO_CLIENT_BUFFER_SIZE = 512
const ECHO_CLIENT_MESSAGE_AMOUNT = 3
const ECHO_CLIENT_MESSAGE_DELAY_MS = 1000

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   int
	InputFile  string
	OutputFile string
}

type Client struct {
	conn   net.Conn
	config ClientConfig
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{conn: conn, config: config}
	return client, nil
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
			continue
		}

		logger.Info(action, logger.Success)
		break
	}

	return conn, err
}

func (client *Client) Run() error {
	const mainAction = "test-echo-server"
	defer client.conn.Close()

	inputFile, err := os.Open(client.config.InputFile)
	if err != nil {
		logger.Error(mainAction, logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(client.config.OutputFile)
	if err != nil {
		logger.Error(mainAction, logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}
	defer outputFile.Close()

	reader := csv.NewReader(inputFile)
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	if err := client.sendBets(reader); err != nil {
		return err
	}

	if err := safe_socket.SendAll(
		client.conn,
		protocol.SerializeFinish(),
	); err != nil {
		return err
	}

	if err := client.receiveWinners(writer); err != nil {
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


func (client *Client) sendBets(reader *csv.Reader) error {
	for {
		row, err := reader.Read()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		newBet, err := client.betFromRow(row)
		if err != nil {
			return err
		}

		packet := protocol.SerializeBet(newBet)

		if err := safe_socket.SendAll(client.conn, packet); err != nil {
			return err
		}
	}
}


func (client *Client) betFromRow(row []string) (*bet.Bet, error) {
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


func (client *Client) receiveWinners(writer *csv.Writer) error {
	for {
		messageType, payload, err := protocol.ReceiveMessage(client.conn)
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
			strconv.Itoa(winner.AgencyId),
			winner.Name,
			winner.LastName,
			strconv.Itoa(winner.Documentation),
			winner.Birthday,
			strconv.Itoa(winner.Number),
		}); err != nil {
			return err
		}
	}
}