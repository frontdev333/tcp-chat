package internal

import (
	"frontdev333/tcp-chat/internal/chat"
	"github.com/google/uuid"
	"net"
	"strconv"
	"time"
)

type Client struct {
	ID       string    `json:"id"`
	Conn     net.Conn  `json:"conn"`
	JoinTime time.Time `json:"join_time"`
	WriteCh  chan chat.ChatMessage
}

func GenerateClientID() string {
	clientIDUint := uint64(uuid.New().ID())
	return strconv.FormatUint(clientIDUint, 10)
}
