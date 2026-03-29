package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type initVals struct {
	timeout *time.Duration

	address string
}

func main() {
	values := getValuesFromFlags()

	client := NewTelnetClient(values.address, *values.timeout, os.Stdin, os.Stdout)

	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	defer client.Close()

	fmt.Fprintf(os.Stderr, "connected to %s ...\n", values.address)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)

	defer cancel()

	sendDone := make(chan error, 1)

	receiveDone := make(chan error, 1)

	go send(client, sendDone)

	go receive(client, receiveDone)

	select {
	case <-ctx.Done():

		// Ctrl+C

		fmt.Fprintln(os.Stderr, "...Interrupted")
	case err := <-sendDone:
		// EOF || err
		if err == nil {
			fmt.Fprintln(os.Stderr, "...EOF")
		} else {
			fmt.Fprintf(os.Stderr, "...Send error: %v\n", err)
		}
	case err := <-receiveDone:
		if err == nil {
			fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
		} else {
			fmt.Fprintf(os.Stderr, "...Receive error: %v\n", err)
		}
	}
}

func getValuesFromFlags() initVals {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")

	flag.Parse()

	if flag.NArg() != 2 {
		log.Fatal("Usage: go-telnet [--timeout=10s] host port")
	}

	host := flag.Arg(0)

	port := flag.Arg(1)

	address := net.JoinHostPort(host, port)

	return initVals{
		address: address,

		timeout: timeout,
	}
}

func send(client TelnetClient, sendDone chan error) {
	for {
		err := client.Send()
		if err != nil {
			sendDone <- err

			return
		}
	}
}

func receive(client TelnetClient, receiveDone chan error) {
	for {
		err := client.Receive()
		if err != nil {
			receiveDone <- err

			return
		}
	}
}
