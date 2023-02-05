package client

import (
	"context"
	"github.com/pkg/errors"
	"net"
	"time"
)

const (
	NetWorkType = "tcp4"
)

type Client struct {
	connection net.Conn
	reader     *Reader
	writer     *Writer
}

func Connect(ctx context.Context, address string, passWord string) (*Client, error) {
	dialer := net.Dialer{
		Timeout:   time.Second * 5,
		KeepAlive: time.Second * 10,
	}

	conn, err := dialer.DialContext(ctx, NetWorkType, address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to %v", address)
	}

	client := &Client{
		connection: conn,
		reader:     NewReader(conn),
		writer:     NewWriter(conn),
	}

	if len(passWord) > 0 {
		_, err = client.Send([]interface{}{
			"AUTH",
			passWord,
		})

		if err != nil {
			return nil, err
		}
	}

	_, err = client.Send([]interface{}{"PING"})
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) Send(values []interface{}) (*Result, error) {
	if err := c.connection.SetDeadline(time.Now().Add(time.Second * 5)); err != nil {
		return nil, err
	}

	if err := c.writer.WriteArray(values); err != nil {
		return nil, errors.Wrapf(err, "failed to execute operation: %v", values[0])
	}

	return c.reader.Read()
}

func (c *Client) Close() error {
	return c.connection.Close()
}
