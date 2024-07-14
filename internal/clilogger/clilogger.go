package clilogger

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

var _ slog.Handler = &Handler{}

type Handler struct {
	output io.Writer
	level  slog.Level
	attrs  []slog.Attr
	groups groups
}

func NewHandler(w io.Writer, level slog.Level) *Handler {
	return &Handler{output: w, level: level}
}

func (h Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := h.clone()
	newHandler.attrs = append(newHandler.attrs, attrs...)
	return newHandler
}

func (h Handler) WithGroup(name string) slog.Handler {
	newHandler := h.clone()
	newHandler.groups = append(h.groups, group{name: name, offset: len(newHandler.attrs)})
	return newHandler
}

func (h Handler) clone() *Handler {
	return &Handler{
		output: h.output,
		level:  h.level,
		attrs:  h.attrs[:len(h.attrs):len(h.attrs)],
		groups: h.groups[:len(h.groups):len(h.groups)],
	}
}

func (h Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h Handler) Handle(_ context.Context, record slog.Record) error {
	var line strings.Builder

	line.WriteString(record.Level.String())
	line.WriteRune(' ')
	line.WriteString(record.Message)
	h.handleAttrs(&line, record)
	line.WriteRune('\n')
	_, err := h.output.Write([]byte(line.String()))
	return err
}

func (h Handler) handleAttrs(line *strings.Builder, record slog.Record) {
	if record.NumAttrs() == 0 && len(h.attrs) == 0 {
		return
	}
	s := make([]string, 0, len(h.attrs)+record.NumAttrs())
	for i, attr := range h.attrs {
		s = append(s, h.groups.prefix(i)+attr.String())
	}
	idx := len(h.attrs)
	record.Attrs(func(attr slog.Attr) bool {
		s = append(s, h.groups.prefix(idx)+attr.String())
		//idx++
		return true
	})
	line.WriteString(" (" + strings.Join(s, ", ") + ")")
}

type group struct {
	name   string
	offset int
}

type groups []group

func (g groups) prefix(index int) string {
	var prefix string
	for _, e := range g {
		if e.offset > index {
			break
		}
		prefix += e.name + "."
	}
	return prefix
}
