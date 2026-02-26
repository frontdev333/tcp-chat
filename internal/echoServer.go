package internal

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"
)

const workersNum = 10

func StartEchoServer(ctx context.Context, port string) error {
	fmt.Println("server started")
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	defer listener.Close()

	jobs := make(chan net.Conn, workersNum)
	defer close(jobs)

	for i := 0; i < workersNum; i++ {
		go handleConn(ctx, jobs)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		select {
		case <-ctx.Done():
			conn.Close()
			return ctx.Err()
		default:
			jobs <- conn
		}
	}
}

func handleConn(ctx context.Context, jobs chan net.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		case conn := <-jobs:
			processConn(ctx, conn)
		}
	}
}

func processConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	defer func(conn net.Conn, b []byte) {
		_, err := conn.Write(b)
		if err != nil {
			slog.Error(err.Error())
		}
	}(conn, []byte("Close connection\n"))

	_, err := conn.Write([]byte("connected to server\n"))
	if err != nil {
		slog.Error(err.Error())
	}

	scanner := bufio.NewScanner(conn)
	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				slog.Error(err.Error())
			}

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if ctx.Err() != nil {
					return
				}
				continue
			}

			return
		}
		msg := scanner.Text() + "\n"
		conn.Write([]byte(msg))
	}
}
