package internal

import (
	"net"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID       string    `json:"id"`
	Conn     net.Conn  `json:"conn"`
	JoinTime time.Time `json:"join_time"`
}

func GenerateClientID() string {
	clientIDUint := uint64(uuid.New().ID())
	return strconv.FormatUint(clientIDUint, 10)
}
