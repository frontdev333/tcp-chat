package domain

import (
	"net"
	"time"
)

type Client struct {
	ID       string    `json:"id"`
	Conn     net.Conn  `json:"conn"`
	JoinTime time.Time `json:"join_time"`
	WriteCh  chan ChatMessage
}
