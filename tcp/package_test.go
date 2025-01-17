package tcp

import (
	"log/slog"
	"testing"
)

func TestMain(m *testing.M) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	m.Run()
}
