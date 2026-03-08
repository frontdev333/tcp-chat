package server

import (
	"context"
	"fmt"
	"frontdev333/tcp-chat/internal/hub"
	"log/slog"
	"net"
)

const workersNum = 100

func StartEchoServer(ctx context.Context, hub *hub.Hub, port string) error {
	fmt.Println("server started")
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	defer listener.Close()

	jobs := make(chan net.Conn, workersNum)
	defer close(jobs)

	for i := 0; i < workersNum; i++ {
		go handleConn(ctx, jobs, hub)
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

func handleConn(ctx context.Context, jobs chan net.Conn, hub *hub.Hub) {
	for {
		select {
		case <-ctx.Done():
			return
		case conn := <-jobs:
			hub.RegisterClient(ctx, conn)
		}
	}
}
