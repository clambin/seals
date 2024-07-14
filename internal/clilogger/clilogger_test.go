package clilogger

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestCLILogger(t *testing.T) {
	tests := []struct {
		name string
		log  func(l *slog.Logger)
		want string
	}{
		{
			name: "no attributes",
			log:  func(l *slog.Logger) { l.Info("Hello world") },
			want: "INFO Hello world\n",
		},
		{
			name: "logged attributes",
			log:  func(l *slog.Logger) { l.Info("Hello world", "attrib", 1) },
			want: "INFO Hello world (attrib=1)\n",
		},
		{
			name: "logger.With()",
			log:  func(l *slog.Logger) { l.With("attrib", 1).Info("Hello world") },
			want: "INFO Hello world (attrib=1)\n",
		},
		{
			name: "combined",
			log:  func(l *slog.Logger) { l.With("attrib", 1).Info("Hello world", "attrib", 2) },
			want: "INFO Hello world (attrib=1, attrib=2)\n",
		},
		{
			name: "logger.WithGroup()",
			log: func(l *slog.Logger) {
				l.
					WithGroup("foo").
					With("attrib", 1).
					WithGroup("bar").
					With("attrib", 2).
					Info("Hello world", "attrib", 3)
			},
			want: "INFO Hello world (foo.attrib=1, foo.bar.attrib=2, foo.bar.attrib=3)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			l := slog.New(NewHandler(&buf, slog.LevelInfo))
			tt.log(l)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func Test_groups(t *testing.T) {
	tests := []struct {
		name   string
		groups groups
		index  int
		want   string
	}{
		{
			name:   "no groups",
			groups: groups{},
			index:  10,
			want:   "",
		},
		{
			name:   "single group",
			groups: groups{{name: "foo", offset: 0}},
			index:  10,
			want:   "foo.",
		},
		{
			name:   "multiple groups - first",
			groups: groups{{name: "foo", offset: 0}, {name: "bar", offset: 2}},
			index:  0,
			want:   "foo.",
		},
		{
			name:   "multiple groups - last",
			groups: groups{{name: "foo", offset: 0}, {name: "bar", offset: 2}},
			index:  2,
			want:   "foo.bar.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.groups.prefix(tt.index))
		})
	}
}
