package main

import (
	"bufio"
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error

	io.Closer

	Send() error

	Receive() error
}

type telnetClient struct {
	address string

	timeout time.Duration

	in io.ReadCloser

	out io.Writer

	conn net.Conn

	inReader *bufio.Reader // для чтения из STDIN

	connReader *bufio.Reader // для чтения из сокета
}

func (tc *telnetClient) Connect() error {
	conn, err := net.DialTimeout("tcp", tc.address, tc.timeout)
	if err != nil {
		return err
	}

	tc.conn = conn

	tc.connReader = bufio.NewReader(conn)

	return nil
}

func (tc *telnetClient) Send() error {
	data, err := tc.inReader.ReadBytes('\n')
	if err != nil {
		return err
	}

	_, err = tc.conn.Write(data)

	return err
}

func (tc *telnetClient) Receive() error {
	data, err := tc.connReader.ReadBytes('\n')
	if err != nil {
		return err
	}

	_, err = tc.out.Write(data)

	return err
}

func (tc *telnetClient) Close() error {
	if tc.conn != nil {
		return tc.conn.Close()
	}

	return nil
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{
		address: address,

		timeout: timeout,

		in: in,

		out: out,

		conn: nil,

		inReader: bufio.NewReader(in),

		connReader: nil, // будет создан в Connect()

	}
}
