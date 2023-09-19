package main

import (
	"bytes"
	"io"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l := listener(t)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			client := NewTelnetClient(l.Addr().String(), timeout(t, "10s"), io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err := client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})
}

func TestActWithClosedClient(t *testing.T) {
	l := listener(t)
	defer func() { require.NoError(t, l.Close()) }()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}

	client := NewTelnetClient(l.Addr().String(), timeout(t, "1s"), io.NopCloser(in), out)
	in.WriteString("hello\n")
	err := client.Send()
	require.Error(t, err)
	require.Equal(t, "send to closed connection", err.Error())

	err = client.Receive()
	require.Error(t, err)
	require.Equal(t, "receive from closed connection", err.Error())

	require.NoError(t, client.Close())
}

func TestPeerClosedConnection(t *testing.T) {
	t.Run("send test", func(t *testing.T) {
		l := listener(t)
		defer func() { l.Close() }()

		in := &bytes.Buffer{}
		out := &bytes.Buffer{}

		client := NewTelnetClient(l.Addr().String(), timeout(t, "1s"), io.NopCloser(in), out)
		require.NoError(t, client.Connect())

		require.NoError(t, l.Close())

		in.WriteString("hello\n")
		err := client.Send()
		require.Error(t, err)
		require.Equal(t, "...Connection was closed by peer", err.Error())

		require.NoError(t, client.Close())
	})

	t.Run("receive test", func(t *testing.T) {
		l := listener(t)
		defer func() { l.Close() }()

		in := &bytes.Buffer{}
		out := &bytes.Buffer{}

		client := NewTelnetClient(l.Addr().String(), timeout(t, "1s"), io.NopCloser(in), out)
		require.NoError(t, client.Connect())

		require.NoError(t, l.Close())
		time.Sleep(time.Second)

		err := client.Receive()
		require.Error(t, err)
		require.Equal(t, "...Connection was closed by peer", err.Error())

		require.NoError(t, client.Close())
	})
}

func TestEOF(t *testing.T) {
	l := listener(t)
	defer func() { require.NoError(t, l.Close()) }()

	in, err := os.OpenFile(createTempFile(t, "empty", ""), os.O_RDONLY, 0755) //nolint:gofumpt
	require.NoError(t, err)
	defer func() { require.NoError(t, in.Close()) }()
	out := &bytes.Buffer{}

	client := NewTelnetClient(l.Addr().String(), timeout(t, "1s"), in, out)
	require.NoError(t, client.Connect())

	err = client.Send()
	require.Error(t, err)
	require.Equal(t, "...EOF", err.Error())

	require.NoError(t, client.Close())
}

func createTempFile(t *testing.T, name string, text string) string {
	t.Helper()

	file, err := os.CreateTemp("", name)
	require.NoError(t, err)

	path := file.Name()
	if len(text) > 0 {
		_, err = file.WriteString(text)
		require.NoError(t, err)
	}
	require.NoError(t, file.Close())

	return path
}

func listener(t *testing.T) net.Listener {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	return l
}

func timeout(t *testing.T, s string) time.Duration {
	t.Helper()
	result, err := time.ParseDuration(s)
	require.NoError(t, err)
	return result
}
