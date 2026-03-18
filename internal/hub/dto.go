package hub

import (
	"frontdev333/tcp-chat/internal"
)

type Request interface {
	execute(map[*internal.Client]bool)
}

type ActiveClientsIDsRequest struct {
	Response chan ActiveClientsIDsResult
}

type ActiveClientsIDsResult []string

type ClientsCountRequest struct {
	Response chan ClientsCountResult
}

type ClientsCountResult int

type ActiveClientsRequest struct {
	Response chan ActiveClientsResult
}

type ActiveClientsResult []*internal.Client

func (r *ActiveClientsIDsRequest) execute(clients map[*internal.Client]bool) {
	res := make([]string, 0, len(clients))

	for c := range clients {
		res = append(res, c.ID)
	}

	r.Response <- res
}

func (r *ClientsCountRequest) execute(clients map[*internal.Client]bool) {
	r.Response <- ClientsCountResult(len(clients))
}

func (a *ActiveClientsRequest) execute(clients map[*internal.Client]bool) {
	res := make([]*internal.Client, len(clients))

	i := 0
	for c, _ := range clients {
		res[i] = c
		i++
	}

	a.Response <- res
}

type ServerStats struct {
	ActiveConnections      int64   `json:"active_connections"`
	TotalMessagesProcessed int64   `json:"total_messages_processed"`
	UptimeSeconds          float64 `json:"uptime_seconds"`
	ErrorCount             int64   `json:"error_count"`
	LastError              string  `json:"last_error"`
	ServerStartTime        string  `json:"server_start_time"`
	MessageRatePerMinute   float64 `json:"message_rate_per_minute"`
}

type HealthCheck struct {
	Status            string  `json:"status"`
	ActiveConnections int64   `json:"active_connections"`
	UptimeSeconds     float64 `json:"uptime_seconds"`
}
