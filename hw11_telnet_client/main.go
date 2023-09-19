package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func sender(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup, t *TelnetClient) {
	defer wg.Done()
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err = (*t).Send()
			if err != nil {
				printStderr(err.Error())
				// Останавливаем соседнюю горутину и закрываем клиента
				stop()
				closeTelnetClient(*t)
				return
			}
		}
	}
}

func receiver(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup, t *TelnetClient) {
	defer wg.Done()
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err = (*t).Receive()
			if err != nil {
				printStderr(err.Error())
				// Останавливаем соседнюю горутину и закрываем клиента
				stop()
				closeTelnetClient(*t)
				return
			}
		}
	}
}

func main() {
	var timeout time.Duration
	flag.DurationVar(
		&timeout,
		"timeout",
		10*time.Second,
		`Connection timeout in "ns", "us" (or "µs"), "ms", "s", "m", "h".)`)
	flag.Parse()

	if flag.NArg() == 0 {
		printStderr(fmt.Sprintf(`No host is defined. Usage: "%s [--timeout=10s] host [port]"`, os.Args[0]))
		return
	}
	addr := net.JoinHostPort(flag.Arg(0), flag.Arg(1))

	t := NewTelnetClient(addr, timeout, os.Stdin, os.Stdout)
	err := t.Connect()
	if err != nil {
		printStderr(err.Error())
		return
	}
	defer closeTelnetClient(t)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	defer stop()

	wg := sync.WaitGroup{}
	wg.Add(2)
	go sender(ctx, stop, &wg, &t)
	go receiver(ctx, stop, &wg, &t)
	wg.Wait()
}

func closeTelnetClient(t TelnetClient) {
	err := t.Close()
	if err != nil {
		log.Fatalln(err)
	}
}

func printStderr(msg string) {
	_, err := fmt.Fprintln(os.Stderr, msg)
	if err != nil {
		log.Fatalln(err)
	}
}
