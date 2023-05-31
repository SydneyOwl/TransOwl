package test

import (
	_ "TransOwl/pkg/logger"
	"github.com/gookit/slog"
	"testing"
)

func TestLogger(t *testing.T) {
	slog.Panicf("testy")
	slog.Warnf("dsd")
}
