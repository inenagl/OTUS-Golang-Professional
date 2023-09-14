package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnetClient struct {
	address  string
	timeout  time.Duration
	inReader *bufio.Reader
	out      io.Writer
	conn     net.Conn
	connScan *bufio.Scanner
	isOpened bool
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{
		address:  address,
		timeout:  timeout,
		inReader: bufio.NewReader(in),
		out:      out,
		isOpened: false,
	}
}

func (t *telnetClient) Connect() error {
	if t.isOpened {
		return nil
	}

	conn, err := net.DialTimeout("tcp", t.address, t.timeout)
	if err != nil {
		return err
	}
	t.conn = conn
	t.connScan = bufio.NewScanner(conn)
	t.isOpened = true

	_, err = fmt.Fprintf(os.Stderr, "...Connected to %s\n", t.address)
	if err != nil {
		return err
	}

	return nil
}

func (t *telnetClient) Close() error {
	if t.isOpened {
		err := t.conn.Close()
		if err != nil {
			return err
		}
		t.isOpened = false
	}

	return nil
}

func (t telnetClient) Send() error {
	if !t.isOpened {
		return errors.New("send to closed connection")
	}

	r := t.inReader
	b, err := r.ReadBytes(byte('\n'))
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = fmt.Errorf("...EOF")
		}
		return err
	}

	if _, err = t.conn.Write(b); err != nil {
		if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
			err = errors.New("...Connection was closed by peer")
		}
		return err
	}

	return nil
}

func (t telnetClient) Receive() error {
	if !t.isOpened {
		return errors.New("receive from closed connection")
	}

	var err error
	s := t.connScan
	if s.Scan() {
		if _, err = t.out.Write(append(s.Bytes(), '\n')); err != nil {
			return err
		}
	}
	if err = s.Err(); err != nil {
		if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
			err = errors.New("...Connection was closed by peer")
		}
		// данная ошибка может возникнуть только в случае, когда сендер закрыл соединение при "ctrl+D"
		// поэтому эту ошибку не выводим.
		if strings.Contains(err.Error(), "use of closed network connection") {
			return nil
		}
		return err
	}

	return nil
}
