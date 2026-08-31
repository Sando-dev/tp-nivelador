package client

import (
	"net"
	"time"
	"os"
	"io"
	"encoding/csv"
	"strconv"
	"strings"

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

	file_input, err := os.Open(client.config.InputFile)
	if err != nil {
		logger.Error(mainAction, logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}
	defer file_input.Close()

	file_output, err := os.Create(client.config.OutputFile)
	if err != nil {
		logger.Error(mainAction, logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}
	defer file_output.Close()

	reader := csv.NewReader(file_input)

	for {
		row, err := reader.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		dni, err := strconv.Atoi(row[2])
		if err != nil {
			logger.Error("parse-dni", logger.Fail, "agency-id", client.config.AgencyId, "row", row)
			return err
		}

		amount, err := strconv.Atoi(row[4])
		if err != nil {
			logger.Error("parse-amount", logger.Fail, "agency-id", client.config.AgencyId, "row", row)
			return err
		}

		newBet := bet.NewBet(client.config.AgencyId, row[0], row[1], dni, row[3], amount)
		packet := protocol.SerializeBet(newBet)

		println("Bet created:", newBet)
		println("Serialized packet:", packet)

		messageArgs := []any{"agency-id", client.config.AgencyId, "message-id", row}

		rowString := strings.Join(row, ",")

		if err := safe_socket.SendAll(client.conn, []byte(rowString)); err != nil {
			logger.Error("send-message", logger.Fail, messageArgs...)
			return err
		}

		responseBuffer, err := safe_socket.RecvAll(client.conn, ECHO_CLIENT_BUFFER_SIZE)
		if err != nil {
			logger.Error("recv-response", logger.Fail, messageArgs...)
			return err
		}

		if string(responseBuffer) != rowString {
			logger.Error("check-response", logger.Fail, messageArgs...)
			return err
		}

		file_output.WriteString(string(responseBuffer) + "\n")

	}

	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)

	return nil
}
