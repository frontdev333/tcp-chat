package internal

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

type Handler struct {
	output io.Writer
	prefix string
	level  slog.Level
}

func NewHandler(output io.Writer, prefix string, level slog.Level) *Handler {
	return &Handler{
		output: output,
		prefix: prefix,
		level:  level,
	}
}

func (h Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h Handler) Handle(_ context.Context, r slog.Record) error {
	var res strings.Builder
	timeStr := r.Time.Format("2006/01/02 15:04:05")
	levelStr := r.Level.String()
	res.WriteString(fmt.Sprintf("[%s] [%s] [%s] [%s]", h.prefix, timeStr, levelStr, r.Message))
	r.Attrs(func(attr slog.Attr) bool {
		res.WriteString(fmt.Sprintf(" %s=%s", attr.Key, attr.Value.String()))
		return true
	})
	res.WriteString(fmt.Sprintln())
	h.output.Write([]byte(res.String()))
	return nil
}

func (h Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h Handler) WithGroup(name string) slog.Handler {
	return h
}
