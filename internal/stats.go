package internal

import "sync/atomic"

type ServerStats struct {
	ActiveConnections      int64        `json:"active_connections"`
	TotalMessagesProcessed int64        `json:"total_messages_processed"`
	UptimeSeconds          float64      `json:"uptime_seconds"`
	ErrorCount             int64        `json:"error_count"`
	LastError              atomic.Value `json:"last_error"`
}
